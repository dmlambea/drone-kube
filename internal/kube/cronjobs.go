package kube

import (
	"github.com/pkg/errors"
	"k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8s "k8s.io/client-go/kubernetes"
)

type batchV1beta1CronJob struct {
}

func (*batchV1beta1CronJob) getObjectMeta(obj runtime.Object) *metav1.ObjectMeta {
	ref := obj.(*v1beta1.CronJob)
	return &ref.ObjectMeta
}

func (*batchV1beta1CronJob) create(cs *k8s.Clientset, obj runtime.Object) (runtime.Object, error) {
	ref := obj.(*v1beta1.CronJob)
	o, err := cs.BatchV1beta1().CronJobs(ref.ObjectMeta.Namespace).Create(ref)
	return o, errors.WithMessagef(err, "unable to create cronjob %s/%s", ref.ObjectMeta.Namespace, ref.ObjectMeta.Name)
}

func (*batchV1beta1CronJob) update(cs *k8s.Clientset, obj runtime.Object) (runtime.Object, error) {
	ref := obj.(*v1beta1.CronJob)
	o, err := cs.BatchV1beta1().CronJobs(ref.ObjectMeta.Namespace).Update(ref)
	return o, errors.WithMessagef(err, "unable to update cronjob %s/%s", ref.ObjectMeta.Namespace, ref.ObjectMeta.Name)
}

func (*batchV1beta1CronJob) find(cs *k8s.Clientset, meta metav1.ObjectMeta) (runtime.Object, error) {
	list, err := cs.BatchV1beta1().CronJobs(meta.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.WithMessagef(err, "unable to find cronjob %s/%s", meta.Namespace, meta.Name)
	}
	for _, ref := range list.Items {
		if ref.ObjectMeta.Name == meta.Name {
			return &ref, nil
		}
	}
	return nil, nil
}
