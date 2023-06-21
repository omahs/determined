package kubernetesrm

import (
	"github.com/sirupsen/logrus"

	k8sV1 "k8s.io/api/core/v1"
	k8Informers "k8s.io/client-go/informers"
	k8sClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type nodeCallbackFunc func(*k8sV1.Node, bool)

type nodeInformer struct {
	informer    k8Informers.SharedInformerFactory
	nodeHandler nodeCallbackFunc
	syslog      *logrus.Entry
	stop        chan struct{}
}

func newNodeInformer(clientSet k8sClient.Interface, nodeHandler nodeCallbackFunc) *nodeInformer {
	return &nodeInformer{
		informer: k8Informers.NewSharedInformerFactoryWithOptions(
			clientSet, 0, []k8Informers.SharedInformerOption{}...),
		nodeHandler: nodeHandler,
		syslog:      logrus.WithField("component", "nodeInformer"),
		stop:        make(chan struct{}),
	}
}

func (n *nodeInformer) startNodeInformer() {
	nodeInformer := n.informer.Core().V1().Nodes().Informer()
	nodeInformer.AddEventHandler(&cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node, ok := obj.(*k8sV1.Node)
			if ok {
				n.syslog.Debugf("node added %s", node.Name)
				go n.nodeHandler(node, true)
			} else {
				n.syslog.Warnf("error converting event of type %T to *k8sV1.Node: %+v", obj, obj)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			node, ok := newObj.(*k8sV1.Node)
			if ok {
				n.syslog.Debugf("node updated %s", node.Name)
				go n.nodeHandler(node, true)
			} else {
				n.syslog.Warnf("error converting event of type %T to *k8sV1.Node: %+v", newObj, newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			node, ok := obj.(*k8sV1.Node)
			if ok {
				n.syslog.Debugf("node stopped %s", node.Name)
				go n.nodeHandler(node, false)
			} else {
				n.syslog.Warnf("error converting event of type %T to *k8sV1.Node: %+v", obj, obj)
			}
		},
	})

	n.syslog.Debug("starting node informer")
	n.informer.Start(n.stop)
	for !nodeInformer.HasSynced() {
	}
	n.syslog.Info("node informer has started")
}

func (n *nodeInformer) stopNodeInformer() {
	n.syslog.Infof("shutting down node informer")
	close(n.stop)
}
