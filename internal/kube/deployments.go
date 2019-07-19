package kube

import (
	"github.com/pkg/errors"
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8s "k8s.io/client-go/kubernetes"
)

type appsV1Deployment struct {
}

func (*appsV1Deployment) getObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	ref := obj.(*v1.Deployment)
	return &ref.ObjectMeta
}

func (*appsV1Deployment) create(cs *k8s.Clientset, obj runtime.Object) (runtime.Object, error) {
	ref := obj.(*v1.Deployment)
	o, err := cs.AppsV1().Deployments(ref.ObjectMeta.Namespace).Create(ref)
	return o, errors.WithMessagef(err, "unable to create deployment %s/%s", ref.ObjectMeta.Namespace, ref.ObjectMeta.Name)
}

func (*appsV1Deployment) update(cs *k8s.Clientset, obj runtime.Object) (runtime.Object, error) {
	ref := obj.(*v1.Deployment)
	o, err := cs.AppsV1().Deployments(ref.ObjectMeta.Namespace).Update(ref)
	return o, errors.WithMessagef(err, "unable to update deployment %s/%s", ref.ObjectMeta.Namespace, ref.ObjectMeta.Name)
}

func (*appsV1Deployment) find(cs *k8s.Clientset, meta metav1.ObjectMeta) (runtime.Object, error) {
	list, err := cs.AppsV1().Deployments(meta.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to find deployment %s/%s", meta.Namespace, meta.Name)
	}
	for _, ref := range list.Items {
		if ref.ObjectMeta.Name == meta.Name {
			return &ref, nil
		}
	}
	return nil, nil
}
