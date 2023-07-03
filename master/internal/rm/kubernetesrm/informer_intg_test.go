package kubernetesrm

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	k8sV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	typedV1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// operations is a tuple struct (name, action) for testing
// events handled by the node informer. Name refers to the
// node name & action refers to the Watch.Event.Type.
type operations struct {
	name   string
	action watch.EventType
}

const namespace = "default"

func initializeMockPods(
	eventChan chan watch.Event,
) *pods {
	mockPodInterface := &mockPodInterface{watcher: &mockWatcher{c: eventChan}}
	return &pods{
		namespace:     namespace,
		podInterfaces: map[string]typedV1.PodInterface{namespace: mockPodInterface},
	}
}

func TestStartInformer(t *testing.T) {
	cases := []struct {
		name            string
		testNamespace   string
		expected        error
		podNames        []string
		orderedPodNames []string
	}{
		{"informer success", namespace,
			nil, []string{"abc"}, []string{"abc"}},
		{"informer success & event ordering success", namespace,
			nil, []string{"A", "B", "C", "D", "E"},
			[]string{"A", "B", "C", "D", "E"}},
		{"informer success & event ordering failure", namespace,
			nil, []string{"A", "B", "C", "D", "E"},
			[]string{"A", "A", "C", "E", "D"}},
		{"podInterface failure", "NOT",
			errors.New("newInformer: passed podInterface is nil"),
			[]string{}, []string{}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			eventChan := make(chan watch.Event)
			ordering := make([]string, 0)
			pods := initializeMockPods(eventChan)
			mockPodHandler := func(pod *k8sV1.Pod) {
				t.Logf("received pod %v", pod)
				ordering = append(ordering, pod.Name)
				wg.Done()
			}

			wg.Add(len(tt.podNames))

			// Test newInformer.
			i, err := newInformer(context.TODO(), tt.testNamespace,
				pods.podInterfaces[tt.testNamespace])
			if err != nil {
				assert.Nil(t, i)
				assert.Error(t, tt.expected, err)
				return

			} else {
				assert.NotNil(t, i)
				assert.Equal(t, tt.expected, err)
			}

			// Test startInformer & assert correct ordering of pod-modified events.
			go i.startInformer(mockPodHandler)
			for _, name := range tt.podNames {
				pod := &k8sV1.Pod{
					ObjectMeta: metaV1.ObjectMeta{
						ResourceVersion: "1",
						Name:            name,
					}}
				eventChan <- watch.Event{
					Type:   watch.Modified,
					Object: pod,
				}
			}
			// Sleep to allow all pod events to be handled.
			time.Sleep(2 * time.Second)
			assert.Equal(t, tt.podNames, ordering)
			wg.Wait()
		})
	}
}

func TestNodeInformer(t *testing.T) {
	cases := []struct {
		name       string
		operations []operations
		output     map[string]bool
		expected   bool
	}{
		{"informer success", []operations{{"abc", watch.Added}},
			map[string]bool{"abc": true}, true},
		{"informer success & event ordering success", []operations{
			{"A", watch.Added}, {"B", watch.Added}, {"C", watch.Added},
			{"A", watch.Deleted}, {"B", watch.Modified}},
			map[string]bool{"B": true, "C": true}, true},
		{"informer success & event ordering failure", []operations{
			{"A", watch.Added}, {"B", watch.Added}, {"C", watch.Added},
			{"A", watch.Deleted}, {"B", watch.Modified}},
			map[string]bool{"A": true, "C": true}, false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			eventChan := make(chan watch.Event)
			currNodes := make(map[string]bool, 0)
			mockNodeInterface := &mockNodeInterface{watcher: &mockWatcher{c: eventChan}}

			wg.Add(len(tt.operations))
			mockNodeHandler := func(node *k8sV1.Node, action watch.EventType) {
				t.Logf("received %v", node.Name)
				if node != nil {
					switch action {
					case watch.Added:
						currNodes[node.Name] = true
					case watch.Modified:
						currNodes[node.Name] = true
					case watch.Deleted:
						delete(currNodes, node.Name)
					default:
						t.Logf("Node did not expect watch.EventType %v", action)
					}
				}
				wg.Done()
			}

			// Test newNodeInformer is created.
			n, err := newNodeInformer(context.TODO(), mockNodeInterface)
			assert.NotNil(t, n)
			assert.Nil(t, err)

			// Test startNodeInformer & iterate through/apply a set of operations
			// (podName, action) to the informer.
			go n.startNodeInformer(mockNodeHandler)
			for _, n := range tt.operations {
				node := &k8sV1.Node{
					ObjectMeta: metaV1.ObjectMeta{
						ResourceVersion: "1",
						Name:            n.name,
					}}
				eventChan <- watch.Event{
					Type:   n.action,
					Object: node,
				}
			}
			// Sleep to allow all node events to be handled.
			time.Sleep(2 * time.Second)

			// Assert equality between expected vs actual status
			// of the pods.
			equality := reflect.DeepEqual(currNodes, tt.output)
			assert.Equal(t, tt.expected, equality)
			wg.Wait()
		})
	}
}
