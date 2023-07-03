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
	nodeInterface typedV1.NodeInterface
	syslog        *logrus.Entry
	resultChan    <-chan watch.Event
}

func newNodeInformer(
	ctx context.Context,
	nodeInterface typedV1.NodeInterface,
) (*nodeInformer, error) {
	nodes, err := nodeInterface.List(ctx, metaV1.ListOptions{LabelSelector: determinedLabel})
	if err != nil {
		return nil, err
	}

	rw, err := watchtools.NewRetryWatcher(nodes.ResourceVersion, &cache.ListWatch{
		WatchFunc: func(options metaV1.ListOptions) (watch.Interface, error) {
			options.LabelSelector = determinedLabel
			return nodeInterface.Watch(ctx, options)
		},
	})
	if err != nil {
		return nil, err
	}

	return &nodeInformer{
		nodeInterface: nodeInterface,
		syslog:        logrus.WithField("component", "nodeInformer"),
		resultChan:    rw.ResultChan(),
	}, nil
}

func (n *nodeInformer) startNodeInformer(nodeHandler nodeCallbackFunc) {
	n.syslog.Info("node informer is starting")
	for event := range n.resultChan {
		if event.Type == watch.Error {
			n.syslog.Warnf("node informer emitted error %+v", event)
			continue
		}

		node, ok := event.Object.(*k8sV1.Node)
		if !ok {
			n.syslog.Warnf("error converting event of type %T to *k8sV1.Node: %+v", event, event)
			continue
		}

		n.syslog.Debugf("informer got new node event for node: %s %s", node.Name, node.Status.Phase)
		nodeHandler(node, event.Type)
	}
	n.syslog.Warn("node informer stopped unexpectedly")
}
