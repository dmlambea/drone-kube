package kube

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s "k8s.io/client-go/kubernetes"
)

var (
	supportedObjectKinds map[schema.GroupVersionKind]applier = map[schema.GroupVersionKind]applier{
		{
			Kind:    "Deployment",
			Group:   "apps",
			Version: "v1",
		}: &appsV1Deployment{},
		{
			Kind:    "CronJob",
			Group:   "batch",
			Version: "v1beta1",
		}: &batchV1beta1CronJob{},
	}
)

// ApplyFunc is the function that performs a kubectl apply-style func for any
// supported object kind. If the object has no namespace, it is assigned the
// given defaultNamespace
type ApplyFunc func(cs *k8s.Clientset, defaultNamespace string, obj runtime.Object) error

// GetApplyFunc returns a valid "apply" function for a given Kubernetes object
// type
func GetApplyFunc(obj runtime.Object) ApplyFunc {
	objKind := obj.GetObjectKind().GroupVersionKind()
	for kind, delegate := range supportedObjectKinds {
		if isKindEquals(kind, objKind) {
			return getApplyFuncForDelegate(delegate)
		}
	}
	return nil
}

type applier interface {
	getObjectMeta(obj runtime.Object) *metav1.ObjectMeta
	create(cs *k8s.Clientset, obj runtime.Object) (runtime.Object, error)
	update(cs *k8s.Clientset, obj runtime.Object) (runtime.Object, error)
	find(cs *k8s.Clientset, meta metav1.ObjectMeta) (runtime.Object, error)
}

func getApplyFuncForDelegate(delegate applier) ApplyFunc {
	return func(cs *k8s.Clientset, defaultNamespace string, obj runtime.Object) error {
		meta := delegate.getObjectMeta(obj)
		if meta.Namespace == "" {
			meta.Namespace = defaultNamespace
		}

		existing, err := delegate.find(cs, *meta)
		if err != nil {
			return err
		}

		// If a current deployment was found, update the provided one, thus
		// overwriting the existing one
		if existing != nil {
			_, err = delegate.update(cs, obj)
			return err
		}

		// Otherwise, create the new deployment since this never existed.
		_, err = delegate.create(cs, obj)
		return err
	}
}

func isKindEquals(k1, k2 schema.GroupVersionKind) bool {
	return k1.Kind == k2.Kind &&
		k1.Group == k2.Group &&
		k1.Version == k2.Version
}
