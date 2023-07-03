package kubernetesrm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	k8sV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

type mockConfigMapInterface struct {
	configMaps map[string]*k8sV1.ConfigMap
	mux        sync.Mutex
}

type mockNodeInterface struct {
	nodes map[string]*k8sV1.Node
	// Simulates latency of the real k8 API server.
	operationalDelay time.Duration
	logMessage       *string
	mux              sync.Mutex
	watcher          *mockWatcher
}

type mockPodInterface struct {
	pods map[string]*k8sV1.Pod
	// Simulates latency of the real k8 API server.
	operationalDelay time.Duration
	logMessage       *string
	mux              sync.Mutex
	watcher          *mockWatcher
}

type mockRoundTripInterface struct {
	message *string
}

type mockWatcher struct {
	c chan watch.Event
}

func (m *mockWatcher) Stop()                          { close(m.c) }
func (m *mockWatcher) ResultChan() <-chan watch.Event { return m.c }

// mockConfig functions
func (m *mockConfigMapInterface) Create(
	ctx context.Context, cm *k8sV1.ConfigMap, opts metaV1.CreateOptions,
) (*k8sV1.ConfigMap, error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, present := m.configMaps[cm.Name]; present {
		return nil, errors.Errorf("configMap with name %s already exists", cm.Name)
	}

	m.configMaps[cm.Name] = cm.DeepCopy()
	return m.configMaps[cm.Name], nil
}

func (m *mockConfigMapInterface) Update(
	context.Context, *k8sV1.ConfigMap, metaV1.UpdateOptions,
) (*k8sV1.ConfigMap, error) {
	panic("implement me")
}

func (m *mockConfigMapInterface) Delete(
	ctx context.Context, name string, options metaV1.DeleteOptions,
) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, present := m.configMaps[name]; !present {
		return errors.Errorf("configMap with name %s doesn't exists", name)
	}

	delete(m.configMaps, name)
	return nil
}

func (m *mockConfigMapInterface) DeleteCollection(
	ctx context.Context, options metaV1.DeleteOptions, listOptions metaV1.ListOptions,
) error {
	panic("implement me")
}

func (m *mockConfigMapInterface) Get(
	ctx context.Context, name string, options metaV1.GetOptions,
) (*k8sV1.ConfigMap, error) {
	panic("implement me")
}

func (m *mockConfigMapInterface) List(
	ctx context.Context, opts metaV1.ListOptions,
) (*k8sV1.ConfigMapList, error) {
	panic("implement me")
}

func (m *mockConfigMapInterface) Watch(
	ctx context.Context, opts metaV1.ListOptions,
) (watch.Interface, error) {
	panic("implement me")
}

func (m *mockConfigMapInterface) Patch(
	ctx context.Context, name string, pt types.PatchType, data []byte, opts metaV1.PatchOptions,
	subresources ...string,
) (result *k8sV1.ConfigMap, err error) {
	panic("implement me")
}

// mockPodInterface functions
func (m *mockPodInterface) Create(
	ctx context.Context, pod *k8sV1.Pod, opts metaV1.CreateOptions,
) (*k8sV1.Pod, error) {
	time.Sleep(m.operationalDelay)
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, present := m.pods[pod.Name]; present {
		return nil, errors.Errorf("pod with name %s already exists", pod.Name)
	}

	m.pods[pod.Name] = pod.DeepCopy()
	return m.pods[pod.Name], nil
}

func (m *mockPodInterface) Update(
	context.Context, *k8sV1.Pod, metaV1.UpdateOptions,
) (*k8sV1.Pod, error) {
	panic("implement me")
}

func (m *mockPodInterface) UpdateStatus(
	context.Context, *k8sV1.Pod, metaV1.UpdateOptions,
) (*k8sV1.Pod, error) {
	panic("implement me")
}

func (m *mockPodInterface) Delete(
	ctx context.Context, name string, options metaV1.DeleteOptions,
) error {
	time.Sleep(m.operationalDelay)
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, present := m.pods[name]; !present {
		return errors.Errorf("pod with name %s doesn't exists", name)
	}

	delete(m.pods, name)
	return nil
}

func (m *mockPodInterface) DeleteCollection(
	ctx context.Context, options metaV1.DeleteOptions, listOptions metaV1.ListOptions,
) error {
	panic("implement me")
}

func (m *mockPodInterface) Get(
	ctx context.Context, name string, options metaV1.GetOptions,
) (*k8sV1.Pod, error) {
	panic("implement me")
}

func (m *mockPodInterface) List(
	ctx context.Context, opts metaV1.ListOptions,
) (*k8sV1.PodList, error) {
	time.Sleep(m.operationalDelay)
	m.mux.Lock()
	defer m.mux.Unlock()

	podList := &k8sV1.PodList{}
	for _, pod := range m.pods {
		podList.Items = append(podList.Items, *pod)
	}
	podList.ResourceVersion = "1"

	return podList, nil
}

func (m *mockPodInterface) Watch(
	ctx context.Context, opts metaV1.ListOptions,
) (watch.Interface, error) {
	if m.watcher == nil {
		return nil, fmt.Errorf("not implemented")
	}
	return m.watcher, nil
}

func (m *mockPodInterface) Patch(
	ctx context.Context, name string, pt types.PatchType, data []byte, opts metaV1.PatchOptions,
	subresources ...string,
) (result *k8sV1.Pod, err error) {
	panic("implement me")
}

func (m *mockPodInterface) GetEphemeralContainers(
	ctx context.Context, podName string, options metaV1.GetOptions,
) (*k8sV1.EphemeralContainers, error) {
	panic("implement me")
}

func (m *mockPodInterface) UpdateEphemeralContainers(
	ctx context.Context, podName string, ephemeralContainers *k8sV1.EphemeralContainers,
	opts metaV1.UpdateOptions,
) (*k8sV1.EphemeralContainers, error) {
	panic("implement me")
}

func (m *mockPodInterface) Bind(context.Context, *k8sV1.Binding, metaV1.CreateOptions) error {
	panic("implement me")
}

func (m *mockPodInterface) Evict(ctx context.Context, eviction *v1beta1.Eviction) error {
	panic("implement me")
}

func (m *mockPodInterface) GetLogs(name string, opts *k8sV1.PodLogOptions) *rest.Request {
	return rest.NewRequestWithClient(&url.URL{}, "", rest.ClientContentConfig{},
		&http.Client{
			Transport: &mockRoundTripInterface{message: m.logMessage},
		})
}

func (m *mockPodInterface) ProxyGet(
	string, string, string, string, map[string]string,
) rest.ResponseWrapper {
	panic("implement me")
}

// mockRoundTripInterface functions
func (m *mockRoundTripInterface) RoundTrip(req *http.Request) (*http.Response, error) {
	var msg string
	if m.message != nil {
		msg = *m.message
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(msg)),
	}, nil
}

// mockNodeInterface functions
func (m *mockNodeInterface) Create(
	ctx context.Context, node *k8sV1.Node, opts metaV1.CreateOptions,
) (*k8sV1.Node, error) {
	time.Sleep(m.operationalDelay)
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, present := m.nodes[node.Name]; present {
		return nil, errors.Errorf("Node with name %s already exists", node.Name)
	}

	m.nodes[node.Name] = node.DeepCopy()
	return m.nodes[node.Name], nil
}

func (m *mockNodeInterface) Update(
	context.Context, *k8sV1.Node, metaV1.UpdateOptions,
) (*k8sV1.Node, error) {
	panic("implement me")
}

func (m *mockNodeInterface) UpdateStatus(
	context.Context, *k8sV1.Node, metaV1.UpdateOptions,
) (*k8sV1.Node, error) {
	panic("implement me")
}

func (m *mockNodeInterface) Delete(
	ctx context.Context, name string, options metaV1.DeleteOptions,
) error {
	time.Sleep(m.operationalDelay)
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, present := m.nodes[name]; !present {
		return errors.Errorf("node with name %s doesn't exists", name)
	}

	delete(m.nodes, name)
	return nil
}

func (m *mockNodeInterface) DeleteCollection(
	ctx context.Context, options metaV1.DeleteOptions, listOptions metaV1.ListOptions,
) error {
	panic("implement me")
}

func (m *mockNodeInterface) Get(
	ctx context.Context, name string, options metaV1.GetOptions,
) (*k8sV1.Node, error) {
	panic("implement me")
}

func (m *mockNodeInterface) List(
	ctx context.Context, opts metaV1.ListOptions,
) (*k8sV1.NodeList, error) {
	time.Sleep(m.operationalDelay)
	m.mux.Lock()
	defer m.mux.Unlock()

	nodeList := &k8sV1.NodeList{}
	for _, node := range m.nodes {
		nodeList.Items = append(nodeList.Items, *node)
	}
	nodeList.ResourceVersion = "1"

	return nodeList, nil
}

func (m *mockNodeInterface) Watch(
	ctx context.Context, opts metaV1.ListOptions,
) (watch.Interface, error) {
	if m.watcher == nil {
		return nil, fmt.Errorf("not implemented")
	}
	return m.watcher, nil
}

func (m *mockNodeInterface) Patch(
	ctx context.Context, name string, pt types.PatchType, data []byte, opts metaV1.PatchOptions,
	subresources ...string,
) (result *k8sV1.Node, err error) {
	panic("implement me")
}

func (m *mockNodeInterface) PatchStatus(ctx context.Context, nodeName string, data []byte) (*v1.Node, error) {
	panic("implement me")
}
