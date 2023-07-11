package kubernetesrm

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8sV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	typedV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
)

type podCallbackFunc func(*k8sV1.Pod)

type informer struct {
	syslog       *logrus.Entry
	podInterface typedV1.PodInterface
	resultChan   <-chan watch.Event
}

func newInformer(
	ctx context.Context,
	namespace string,
	podInterface typedV1.PodInterface,
) (*informer, error) {
	if podInterface == nil {
		return nil, errors.New("newInformer: passed podInterface is nil")
	}

	pods, err := podInterface.List(ctx, metaV1.ListOptions{LabelSelector: determinedLabel})
	if err != nil {
		return nil, err
	}

	rw, err := watchtools.NewRetryWatcher(pods.ResourceVersion, &cache.ListWatch{
		WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
			options.LabelSelector = determinedLabel
			return podInterface.Watch(ctx, options)
		},
	})
	if err != nil {
		return nil, err
	}

	return &informer{
		syslog:       logrus.WithField("PodInformer", namespace),
		podInterface: podInterface,
		resultChan:   rw.ResultChan(),
	}, nil
}

// startInformer returns the updated pod, if any.
func (i *informer) startInformer(cb podCallbackFunc) {
	i.syslog.Info("pod informer is starting")
	for event := range i.resultChan {
		if event.Type == watch.Error {
			i.syslog.Warnf("pod informer emitted error %+v", event)
			continue
		}

		pod, ok := event.Object.(*k8sV1.Pod)
		if !ok {
			i.syslog.Warnf("error converting event of type %T to *k8sV1.Pod: %+v", event, event)
			continue
		}

		i.syslog.Debugf("informer got new pod event for pod: %s %s", pod.Name, pod.Status.Phase)
		cb(pod)
	}
	i.syslog.Warn("pod informer stopped unexpectedly")
}
