//go:build integration
// +build integration

package kubernetesrm

import (
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	k8sV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/determined-ai/determined/master/internal/config"
	"github.com/determined-ai/determined/master/internal/mocks"
	"github.com/determined-ai/determined/master/pkg/device"
	"github.com/determined-ai/determined/master/pkg/model"
)

const (
	defaultNamespace    = "default"
	auxNodeName         = "aux"
	compNodeName        = "comp"
	pod1NumSlots        = 4
	pod2NumSlots        = 8
	nodeNumSlots        = 8
	nodeNumSlotsCPU     = 20
	podServSlotTypeGPU  = "randomDefault"
	cpuResourceRequests = int64(4000)
)

// Number of active slots on given nodes.
var nodeToSlots = map[string]int{
	auxNodeName:  pod1NumSlots,
	compNodeName: pod2NumSlots,
}

func TestGetAgentsAndGetSlots(t *testing.T) {
	// Create resource list.
	resourceList := map[k8sV1.ResourceName]resource.Quantity{
		k8sV1.ResourceName(ResourceTypeNvidia): *resource.NewQuantity(
			int64(nodeNumSlots),
			resource.Format("DecimalSI")),
		k8sV1.ResourceCPU: *resource.NewQuantity(
			int64(nodeNumSlotsCPU),
			resource.Format("DecimalSI")),
	}

	// Create auxiliary agent node and compute node.
	auxNode := k8sV1.Node{
		ObjectMeta: metaV1.ObjectMeta{
			ResourceVersion: "1",
			Name:            auxNodeName,
		},
		Status: k8sV1.NodeStatus{Allocatable: resourceList},
	}

	compNode := k8sV1.Node{
		ObjectMeta: metaV1.ObjectMeta{
			ResourceVersion: "1",
			Name:            compNodeName,
		},
		Status: k8sV1.NodeStatus{Allocatable: resourceList},
	}

	nodes := map[string]*k8sV1.Node{
		auxNode.Name:  &auxNode,
		compNode.Name: &compNode,
	}

	type AgentsTestCase struct {
		agentsName   string
		podsService  *pods
		wantedAgents map[string]int
		agentName    string
		slotsName    string
		slotName     string
	}
	tests := []AgentsTestCase{
		{
			agentsName:   "GPUPodService",
			podsService:  createMockPodsService(nodes, podServSlotTypeGPU),
			wantedAgents: map[string]int{auxNodeName: 0, compNodeName: 0},
			agentName:    "agent",
			slotsName:    "slots",
			slotName:     "slot",
		},
		{
			agentsName:   "CUDAPodService",
			podsService:  createMockPodsService(nodes, device.CUDA),
			wantedAgents: map[string]int{auxNodeName: 0, compNodeName: 0},
			agentName:    "agent",
			slotsName:    "slots",
			slotName:     "slot",
		},
		{
			agentsName:   "CPUPodService",
			podsService:  createMockPodsService(nodes, device.CPU),
			wantedAgents: map[string]int{auxNodeName: 0, compNodeName: 0},
			agentName:    "agent",
			slotsName:    "slots",
			slotName:     "slot",
		},
	}

	for _, test := range tests {
		t.Run(test.agentsName, func(t *testing.T) {
			agentsResp := test.podsService.handleGetAgentsRequest()
			require.Equal(t, len(test.wantedAgents), len(agentsResp.Agents))
			for _, agent := range agentsResp.Agents {
				_, ok := test.wantedAgents[agent.Id]
				require.True(t, ok)
				agentTestName := test.agentName + "_" + agent.Id
				t.Run(agentTestName, func(t *testing.T) {
					agentResp := test.podsService.handleGetAgentRequest(agent.Id)
					require.Equal(t, agentResp.Agent.Id, agent.Id)

					t.Run(test.slotsName, func(t *testing.T) {
						slotsResp := test.podsService.handleGetSlotsRequest(agent.Id)
						require.Equal(t, nodeNumSlots, len(slotsResp.Slots))

						activeSlots := 0
						cnt := 0
						for i, slot := range slotsResp.Slots {
							slotTestName := test.slotName + "_" + strconv.Itoa(i)
							t.Run(slotTestName, func(t *testing.T) {
								slotResp := test.podsService.handleGetSlotRequest(agent.Id,
									strconv.Itoa(cnt))
								slotID, err := strconv.Atoi(slotResp.Slot.Id)
								require.NoError(t, err)
								require.Equal(t, cnt, slotID)
								if slot.Container != nil {
									activeSlots++
									require.True(t, slotResp.Slot.Container != nil)
								}
							})
						}
						require.Equal(t, nodeToSlots[agent.Id], activeSlots)
					})
				})
			}
		})
	}
}

// At the moment, ROCm resource manager should panic when this device type is used in requests
// about cluster resources because ROCm is currently not supported in Kubernetes, so we check
// this resource manager separately.
func TestROCmPodsService(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func()
	}{
		{name: "GetAgentsROCM", testFunc: testROCMGetAgents},
		{name: "GetAgentROCM", testFunc: testROCMGetAgent},
		{name: "GetSlotsROCM", testFunc: testROCMGetSlots},
		{name: "GetSlotROCM", testFunc: testROCMGetSlot},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) { require.Panics(t, test.testFunc) })
	}
}

func testROCMGetAgents() {
	ps := createMockPodsService(createAuxNodeMap(), device.ROCM)
	ps.handleGetAgentsRequest()
}

func testROCMGetAgent() {
	nodes := createAuxNodeMap()
	ps := createMockPodsService(nodes, device.ROCM)
	ps.handleGetAgentRequest(auxNodeName)
}

func testROCMGetSlots() {
	nodes := createAuxNodeMap()
	ps := createMockPodsService(nodes, device.ROCM)
	ps.handleGetSlotsRequest(auxNodeName)
}

func testROCMGetSlot() {
	nodes := createAuxNodeMap()
	ps := createMockPodsService(nodes, device.ROCM)
	for i := 0; i < nodeNumSlots; i++ {
		ps.handleGetSlotRequest(auxNodeName, strconv.Itoa(i))
	}
}

func createAuxNodeMap() map[string]*k8sV1.Node {
	// Create resource list and auxiliary agent node.
	resourceList := map[k8sV1.ResourceName]resource.Quantity{
		k8sV1.ResourceName(ResourceTypeNvidia): *resource.NewQuantity(
			int64(nodeNumSlots),
			resource.Format("DecimalSI")),
		k8sV1.ResourceCPU: *resource.NewQuantity(
			int64(nodeNumSlotsCPU),
			resource.Format("DecimalSI")),
	}

	auxNode := k8sV1.Node{
		ObjectMeta: metaV1.ObjectMeta{
			ResourceVersion: "1",
			Name:            auxNodeName,
		},
		Status: k8sV1.NodeStatus{Allocatable: resourceList},
	}
	return map[string]*k8sV1.Node{
		auxNode.Name: &auxNode,
	}
}

// createMockPodsService creates two pods. One pod is run on the auxiliary node and the other is
// run on the compute node.
func createMockPodsService(nodes map[string]*k8sV1.Node, devSlotType device.Type) *pods {
	// Create pods.
	pod1 := &pod{
		allocationID: model.AllocationID(uuid.New().String()),
		slots:        pod1NumSlots,
		pod: &k8sV1.Pod{
			Spec: k8sV1.PodSpec{NodeName: auxNodeName},
		},
	}
	pod2 := &pod{
		allocationID: model.AllocationID(uuid.New().String()),
		slots:        pod2NumSlots,
		pod: &k8sV1.Pod{
			Spec: k8sV1.PodSpec{NodeName: compNodeName},
		},
	}

	podHandlers := map[string]*pod{
		string(pod1.allocationID): pod1,
		string(pod2.allocationID): pod2,
	}

	// Create pod service client set.
	podsClientSet := &mocks.K8sClientsetInterface{}
	coreV1Interface := &mocks.K8sCoreV1Interface{}
	podsInterface := &mocks.PodInterface{}
	pList := &k8sV1.PodList{Items: []k8sV1.Pod{}}
	podsInterface.On("List", mock.Anything, mock.Anything).Return(pList, nil)
	coreV1Interface.On("Pods", mock.Anything).Return(podsInterface)
	podsClientSet.On("CoreV1").Return(coreV1Interface)

	return &pods{
		namespace:           defaultNamespace,
		namespaceToPoolName: make(map[string]string),
		currentNodes:        nodes,
		podNameToPodHandler: podHandlers,
		slotType:            devSlotType,
		syslog:              logrus.WithField("namespace", namespace),
		nodeToSystemResourceRequests: map[string]int64{
			auxNodeName:  cpuResourceRequests,
			compNodeName: cpuResourceRequests,
		},
		slotResourceRequests: config.PodSlotResourceRequests{CPU: 2},
		clientSet:            podsClientSet,
	}
}
