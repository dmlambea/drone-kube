package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"

	"github.com/pkg/errors"

	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

)

type (
	repo struct {
		Owner string
		Name  string
	}

	build struct {
		Tag     string
		Event   string
		Number  int
		Commit  string
		Ref     string
		Branch  string
		Author  string
		Status  string
		Link    string
		Started int64
		Created int64
	}

	job struct {
		Started int64
	}

	config struct {
		Kubeconfig string
		Ca        string
		Server    string
		Token     string
		Namespace string
		Template  string
	}

	plugin struct {
		Repo   repo
		Build  build
		Config config
		Job    job
		cs *k8s.Clientset
	}
)

func (p plugin) exec() (err error) {
	if err = p.initPlugin(); err != nil {
		return errors.WithMessage(err, "initialization failed")
	}

	var dep v1.Deployment
	if dep, err = p.makeDeploymentDescriptor(); err != nil {
		return errors.WithMessage(err, "unable to make deployment")
	}

	if err = p.updateOrCreateDeployment(dep); err != nil {
		return errors.WithMessage(err, "unable to apply deployment")
	}
	return
}

func (p *plugin) initPlugin() (err error) {
	if err = p.checkConfig(); err != nil {
		return errors.WithMessage(err, "configuration error")
	}
	if p.Config.Namespace == "" {
		p.Config.Namespace = "default"
	}

	var clientCfg *rest.Config
	if clientCfg, err = getClientConfig(p.Config); err != nil {
		return errors.WithMessage(err, "unable to get client config")
	}

	var clientSet *k8s.Clientset
	if clientSet, err = kubernetes.NewForConfig(clientCfg); err != nil {
		return errors.WithMessage(err, "unable to create Kubernetes client")
	}
	p.cs = clientSet
	return nil
}

func (p plugin) makeDeploymentDescriptor() (deployment v1.Deployment, err error) {
	var txt string
	if txt, err = openAndSub(p.Config.Template, p); err != nil {
		return deployment, errors.WithMessage(err, "template processing failed")
	}
	
	decode := scheme.Codecs.UniversalDeserializer().Decode
	var obj runtime.Object
	if obj, _, err = decode([]byte(txt), nil, nil); err != nil {
		return deployment, errors.WithMessagef(err, "unable to deserialize deployment: %v", txt)
	}

	dep := obj.(*v1.Deployment)
	deployment = *dep
	if deployment.ObjectMeta.Namespace == "" {
		deployment.ObjectMeta.Namespace = "default"
	}
	return
}

func (p plugin) updateOrCreateDeployment(dep v1.Deployment) (err error) {
	var currentDep v1.Deployment
	if currentDep, err = findDeployment(dep.ObjectMeta.Name, dep.ObjectMeta.Namespace, p.cs); err != nil {
		return errors.WithMessagef(err, "unable to find deployment %s/%s", dep.ObjectMeta.Namespace, dep.ObjectMeta.Name)
	}

	if currentDep.ObjectMeta.Name == dep.ObjectMeta.Name {
		// update the existing deployment, ignore the deployment that it comes back with
		_, err = p.cs.AppsV1().Deployments(p.Config.Namespace).Update(&dep)
		return errors.WithMessagef(err, "unable to update deployment %s/%s", dep.ObjectMeta.Namespace, dep.ObjectMeta.Name)
	}

	// create the new deployment since this never existed.
	_, err = p.cs.AppsV1().Deployments(p.Config.Namespace).Create(&dep)
	return errors.WithMessagef(err, "unable to create deployment %s/%s", dep.ObjectMeta.Namespace, dep.ObjectMeta.Name)
}

func (p plugin) checkConfig() error {
	if p.Config.Template == "" {
		return errors.New("KUBE_TEMPLATE, or template must be defined")
	}
	if p.Config.Kubeconfig != "" {
		return nil
	}

	if p.Config.Server == "" {
		return errors.New("if not using KUBE_CONFIG, KUBE_SERVER must be defined")
	}
	if p.Config.Token == "" {
		return errors.New("if not using KUBE_CONFIG, KUBE_TOKEN must be defined")
	}
	if p.Config.Ca == "" {
		return errors.New("if not using KUBE_CONFIG, KUBE_CA must be defined")
	}
	return nil
}

func getClientConfig(cfg config) (*rest.Config, error) {
	if cfg.Kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	}

	ca, err := base64.StdEncoding.DecodeString(cfg.Ca)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to base64-decode CA")
	}
	defaultConfig := api.NewConfig()
	defaultConfig.Clusters["drone"] = &api.Cluster{
		Server:                   cfg.Server,
		CertificateAuthorityData: ca,
	}
	defaultConfig.AuthInfos["drone"] = &api.AuthInfo{
		Token: cfg.Token,
	}

	defaultConfig.Contexts["drone"] = &api.Context{
		Cluster:  "drone",
		AuthInfo: "drone",
	}
	defaultConfig.CurrentContext = "drone"

	clientBuilder := clientcmd.NewNonInteractiveClientConfig(*defaultConfig, "drone", &clientcmd.ConfigOverrides{}, nil)
	return clientBuilder.ClientConfig()
}

func findDeployment(depName string, namespace string, cs *k8s.Clientset) (v1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}
	var d v1.Deployment
	deployments, err := listDeployments(cs, namespace)
	if err != nil {
		return d, err
	}
	for _, thisDep := range deployments {
		if thisDep.ObjectMeta.Name == depName {
			return thisDep, err
		}
	}
	return d, err
}

// List the deployments
func listDeployments(cs *k8s.Clientset, namespace string) ([]v1.Deployment, error) {
	// docs on this:
	// https://github.com/kubernetes/client-go/blob/master/pkg/apis/extensions/types.go
	deployments, err := cs.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
	}
	return deployments.Items, err
}

// open up the template and then sub variables in. Handlebar stuff.
func openAndSub(templateFile string, p plugin) (string, error) {
	t, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return "", err
	}
	return RenderTrim(string(t), p)
}
