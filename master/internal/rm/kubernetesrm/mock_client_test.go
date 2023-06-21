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
	"k8s.io/api/policy/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	discovery "k8s.io/client-go/discovery"
	admissionregistrationv1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	admissionregistrationv1beta1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1beta1"
	internalv1alpha1 "k8s.io/client-go/kubernetes/typed/apiserverinternal/v1alpha1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	appsv1beta1 "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	appsv1beta2 "k8s.io/client-go/kubernetes/typed/apps/v1beta2"
	authenticationv1 "k8s.io/client-go/kubernetes/typed/authentication/v1"
	authenticationv1beta1 "k8s.io/client-go/kubernetes/typed/authentication/v1beta1"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	authorizationv1beta1 "k8s.io/client-go/kubernetes/typed/authorization/v1beta1"
	autoscalingv1 "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
	autoscalingv2beta1 "k8s.io/client-go/kubernetes/typed/autoscaling/v2beta1"
	autoscalingv2beta2 "k8s.io/client-go/kubernetes/typed/autoscaling/v2beta2"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	batchv1beta1 "k8s.io/client-go/kubernetes/typed/batch/v1beta1"
	batchv2alpha1 "k8s.io/client-go/kubernetes/typed/batch/v2alpha1"
	certificatesv1 "k8s.io/client-go/kubernetes/typed/certificates/v1"
	certificatesv1beta1 "k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
	coordinationv1 "k8s.io/client-go/kubernetes/typed/coordination/v1"
	coordinationv1beta1 "k8s.io/client-go/kubernetes/typed/coordination/v1beta1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	discoveryv1alpha1 "k8s.io/client-go/kubernetes/typed/discovery/v1alpha1"
	discoveryv1beta1 "k8s.io/client-go/kubernetes/typed/discovery/v1beta1"
	eventsv1 "k8s.io/client-go/kubernetes/typed/events/v1"
	eventsv1beta1 "k8s.io/client-go/kubernetes/typed/events/v1beta1"
	extensionsv1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	flowcontrolv1alpha1 "k8s.io/client-go/kubernetes/typed/flowcontrol/v1alpha1"
	flowcontrolv1beta1 "k8s.io/client-go/kubernetes/typed/flowcontrol/v1beta1"
	networkingv1 "k8s.io/client-go/kubernetes/typed/networking/v1"
	networkingv1beta1 "k8s.io/client-go/kubernetes/typed/networking/v1beta1"
	nodev1 "k8s.io/client-go/kubernetes/typed/node/v1"
	nodev1alpha1 "k8s.io/client-go/kubernetes/typed/node/v1alpha1"
	nodev1beta1 "k8s.io/client-go/kubernetes/typed/node/v1beta1"
	policyv1beta1 "k8s.io/client-go/kubernetes/typed/policy/v1beta1"
	rbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	rbacv1alpha1 "k8s.io/client-go/kubernetes/typed/rbac/v1alpha1"
	rbacv1beta1 "k8s.io/client-go/kubernetes/typed/rbac/v1beta1"
	schedulingv1 "k8s.io/client-go/kubernetes/typed/scheduling/v1"
	schedulingv1alpha1 "k8s.io/client-go/kubernetes/typed/scheduling/v1alpha1"
	schedulingv1beta1 "k8s.io/client-go/kubernetes/typed/scheduling/v1beta1"
	storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
	storagev1alpha1 "k8s.io/client-go/kubernetes/typed/storage/v1alpha1"
	storagev1beta1 "k8s.io/client-go/kubernetes/typed/storage/v1beta1"
	rest "k8s.io/client-go/rest"
)

type mockConfigMapInterface struct {
	configMaps map[string]*k8sV1.ConfigMap
	mux        sync.Mutex
}

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

type mockPodInterface struct {
	pods map[string]*k8sV1.Pod
	// Simulates latency of the real k8 API server.
	operationalDelay time.Duration
	logMessage       *string
	mux              sync.Mutex
	watcher          *mockWatcher
}

type mockWatcher struct {
	c chan watch.Event
}

func (m *mockWatcher) Stop()                          { close(m.c) }
func (m *mockWatcher) ResultChan() <-chan watch.Event { return m.c }

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

type mockRoundTripInterface struct {
	message *string
}

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

type mockClientSet struct {
	coreV1  corev1.CoreV1Interface
	watcher *mockWatcher
}

func (m *mockClientSet) Discovery() discovery.DiscoveryInterface {
	panic("implement me")
}
func (m *mockClientSet) AdmissionregistrationV1() admissionregistrationv1.AdmissionregistrationV1Interface {
	panic("implement me")
}
func (m *mockClientSet) AdmissionregistrationV1beta1() admissionregistrationv1beta1.AdmissionregistrationV1beta1Interface {
	panic("implement me")
}

func (m *mockClientSet) InternalV1alpha1() internalv1alpha1.InternalV1alpha1Interface {
	panic("implement me")
}
func (m *mockClientSet) AppsV1() appsv1.AppsV1Interface {
	panic("implement me")
}
func (m *mockClientSet) AppsV1beta1() appsv1beta1.AppsV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) AppsV1beta2() appsv1beta2.AppsV1beta2Interface {
	panic("implement me")
}
func (m *mockClientSet) AuthenticationV1() authenticationv1.AuthenticationV1Interface {
	panic("implement me")
}
func (m *mockClientSet) AuthenticationV1beta1() authenticationv1beta1.AuthenticationV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) AuthorizationV1() authorizationv1.AuthorizationV1Interface {
	panic("implement me")
}
func (m *mockClientSet) AuthorizationV1beta1() authorizationv1beta1.AuthorizationV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) AutoscalingV1() autoscalingv1.AutoscalingV1Interface {
	panic("implement me")
}
func (m *mockClientSet) AutoscalingV2beta1() autoscalingv2beta1.AutoscalingV2beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) AutoscalingV2beta2() autoscalingv2beta2.AutoscalingV2beta2Interface {
	panic("implement me")
}
func (m *mockClientSet) BatchV1() batchv1.BatchV1Interface {
	panic("implement me")
}
func (m *mockClientSet) BatchV1beta1() batchv1beta1.BatchV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) BatchV2alpha1() batchv2alpha1.BatchV2alpha1Interface {
	panic("implement me")
}
func (m *mockClientSet) CertificatesV1() certificatesv1.CertificatesV1Interface {
	panic("implement me")
}
func (m *mockClientSet) CertificatesV1beta1() certificatesv1beta1.CertificatesV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) CoordinationV1beta1() coordinationv1beta1.CoordinationV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) CoordinationV1() coordinationv1.CoordinationV1Interface {
	panic("implement me")
}
func (m *mockClientSet) CoreV1() corev1.CoreV1Interface {
	// panic("implement me")
	return m.coreV1
}
func (m *mockClientSet) DiscoveryV1alpha1() discoveryv1alpha1.DiscoveryV1alpha1Interface {
	panic("implement me")
}
func (m *mockClientSet) DiscoveryV1beta1() discoveryv1beta1.DiscoveryV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) EventsV1() eventsv1.EventsV1Interface {
	panic("implement me")
}
func (m *mockClientSet) EventsV1beta1() eventsv1beta1.EventsV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) ExtensionsV1beta1() extensionsv1beta1.ExtensionsV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) FlowcontrolV1alpha1() flowcontrolv1alpha1.FlowcontrolV1alpha1Interface {
	panic("implement me")
}
func (m *mockClientSet) FlowcontrolV1beta1() flowcontrolv1beta1.FlowcontrolV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) NetworkingV1() networkingv1.NetworkingV1Interface {
	panic("implement me")
}
func (m *mockClientSet) NetworkingV1beta1() networkingv1beta1.NetworkingV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) NodeV1() nodev1.NodeV1Interface {
	panic("implement me")
}
func (m *mockClientSet) NodeV1alpha1() nodev1alpha1.NodeV1alpha1Interface {
	panic("implement me")
}
func (m *mockClientSet) NodeV1beta1() nodev1beta1.NodeV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) PolicyV1beta1() policyv1beta1.PolicyV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) RbacV1() rbacv1.RbacV1Interface {
	panic("implement me")
}
func (m *mockClientSet) RbacV1beta1() rbacv1beta1.RbacV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) RbacV1alpha1() rbacv1alpha1.RbacV1alpha1Interface {
	panic("implement me")
}
func (m *mockClientSet) SchedulingV1alpha1() schedulingv1alpha1.SchedulingV1alpha1Interface {
	panic("implement me")
}
func (m *mockClientSet) SchedulingV1beta1() schedulingv1beta1.SchedulingV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) SchedulingV1() schedulingv1.SchedulingV1Interface {
	panic("implement me")
}
func (m *mockClientSet) StorageV1beta1() storagev1beta1.StorageV1beta1Interface {
	panic("implement me")
}
func (m *mockClientSet) StorageV1() storagev1.StorageV1Interface {
	panic("implement me")
}
func (m *mockClientSet) StorageV1alpha1() storagev1alpha1.StorageV1alpha1Interface {
	panic("implement me")
}
