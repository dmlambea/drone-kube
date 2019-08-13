package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/dmlambea/drone-kube/internal/kube"
)

const defaultCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

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
		ClientCert string
		ClientKey  string
		Ca         string
		Server     string
		Token      string
		Namespace  string
		Template   string
	}

	plugin struct {
		Repo   repo
		Build  build
		Config config
		Job    job
		cs     *k8s.Clientset
	}
)

func (p plugin) exec() (err error) {
	if err = p.initPlugin(); err != nil {
		return errors.WithMessage(err, "initialization failed")
	}

	var objs []runtime.Object
	if objs, err = p.makeObjectDescriptors(); err != nil {
		return errors.WithMessage(err, "unable to make object descriptors")
	}

	for idx, o := range objs {
		if err = p.updateOrCreateObject(o); err != nil {
			return errors.WithMessagef(err, "unable to apply descriptor %d", idx)
		}
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

func (p plugin) makeObjectDescriptors() (obj []runtime.Object, err error) {
	var txt string
	if txt, err = openAndSub(p.Config.Template, p); err != nil {
		err = errors.WithMessage(err, "template processing failed")
		return
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	chunks := regexp.MustCompile("---[[:space:]]*\n").Split(txt, -1)
	fmt.Printf("%d documents found\n", len(chunks))
	for idx, c := range chunks {
		fmt.Printf(" - Parsing document %d\n", idx)
		var o runtime.Object
		if o, _, err = decode([]byte(c), nil, nil); err != nil {
			err = errors.WithMessagef(err, "unable to deserialize object %d: %v", idx, c)
			break
		}
		obj = append(obj, o)
	}
	return
}

func (p plugin) updateOrCreateObject(obj runtime.Object) (err error) {
	k := obj.GetObjectKind().GroupVersionKind()
	fmt.Printf("Creating/updating object %s %s/%s\n", k.Kind, k.Group, k.Version)

	if f := kube.GetApplyFunc(obj); f != nil {
		return f(p.cs, p.Config.Namespace, obj)
	}
	return errors.Errorf("unsupported object kind %s %s/%s", k.Kind, k.Group, k.Version)
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
	if p.Config.Ca == "" {
		if _, err := os.Stat(defaultCAFile); os.IsNotExist(err) {
			return errors.Errorf("if not using KUBE_CONFIG and no default CA file %s exists, KUBE_CA must be defined", defaultCAFile)
		}
	}
	if p.Config.ClientCert != "" && p.Config.ClientKey != "" {
		return nil
	}
	if p.Config.Token == "" {
		return errors.New("no KUBE_CONFIG, client certificate nor KUBE_TOKEN defined")
	}
	return nil
}

func getClientConfig(cfg config) (*rest.Config, error) {
	if cfg.Kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	}

	defaultConfig := api.NewConfig()
	defaultConfig.Clusters["drone"] = &api.Cluster{
		Server:               cfg.Server,
		CertificateAuthority: defaultCAFile,
	}
	if cfg.Ca != "" {
		ca, err := base64.StdEncoding.DecodeString(cfg.Ca)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to base64-decode CA")
		}
		defaultConfig.Clusters["drone"].CertificateAuthorityData = ca
	}

	authInfo := api.AuthInfo{
		Token:             cfg.Token,
		ClientCertificate: cfg.ClientCert,
		ClientKey:         cfg.ClientKey,
	}
	defaultConfig.AuthInfos["drone"] = &authInfo
	defaultConfig.Contexts["drone"] = &api.Context{
		Cluster:  "drone",
		AuthInfo: "drone",
	}
	defaultConfig.CurrentContext = "drone"

	clientBuilder := clientcmd.NewNonInteractiveClientConfig(*defaultConfig, "drone", &clientcmd.ConfigOverrides{}, nil)
	return clientBuilder.ClientConfig()
}

// open up the template and then sub variables in. Handlebar stuff.
func openAndSub(templateFile string, p plugin) (string, error) {
	t, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return "", err
	}
	return RenderTrim(string(t), p)
}
