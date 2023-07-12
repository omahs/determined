package kubernetesrm

import (
	"context"

	"github.com/sirupsen/logrus"
	k8sV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	typedV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
)

type nodeCallbackFunc func(*k8sV1.Node, watch.EventType)

type nodeInformer struct {
	cb            nodeCallbackFunc
	nodeInterface typedV1.NodeInterface
	syslog        *logrus.Entry
	resultChan    <-chan watch.Event
	doneChan      <-chan struct{}
}

func newNodeInformer(
	ctx context.Context,
	nodeInterface typedV1.NodeInterface,
	cb nodeCallbackFunc,
) (*nodeInformer, error) {
	nodes, err := nodeInterface.List(ctx, metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}

	rw, err := watchtools.NewRetryWatcher(nodes.ResourceVersion, &cache.ListWatch{
		WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
			return nodeInterface.Watch(ctx, options)
		},
	})
	if err != nil {
		return nil, err
	}
	for _, node := range nodes.Items {
		cb(&node, watch.Added)
	}
	return &nodeInformer{
		cb:            cb,
		nodeInterface: nodeInterface,
		syslog:        logrus.WithField("component", "nodeInformer"),
		resultChan:    rw.ResultChan(),
		doneChan:      ctx.Done(),
	}, nil
}

func (n *nodeInformer) startNodeInformer() {
	n.syslog.Info("node informer is starting")
	defer n.syslog.Warn("node informer stopped unexpectedly")

	for {
		select {
		case <-n.doneChan:
			return
		case event := <-n.resultChan:
			if event.Type == watch.Error {
				n.syslog.Warnf("node informer emitted error %+v", event)
				continue
			}

			node, ok := event.Object.(*k8sV1.Node)
			if !ok {
				n.syslog.Warnf("error converting event of type %T to *k8sV1.Node: %+v", event, event)
				continue
			}

			n.syslog.Debugf("informer got new node event(%s) for node: %s %s", event.Type, node.Name, node.Status.Phase)
			n.cb(node, event.Type)
		}
	}
}
