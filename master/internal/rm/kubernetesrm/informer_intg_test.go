package kubernetesrm

import (
	"context"
	"fmt"
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

type operations struct {
	name   string
	action string
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
		name          string
		testNamespace string
		expected      error
		podNames      []string
	}{
		{"informer success", namespace,
			nil, []string{"abc"}},
		{"informer success & event ordering", namespace,
			nil, []string{"A", "B", "C", "D", "E"}},
		{"podInterface failure", "NOT",
			errors.New("newInformer: passed podInterface is nil"), []string{}},
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
			time.Sleep(3 * time.Second)
			assert.Equal(t, tt.podNames, ordering)
			wg.Wait()
		})
	}
}

func TestNodeInformer(t *testing.T) {
	cases := []struct {
		name          string
		testNamespace string
		expected      error
		operations    []operations
		output        map[string]bool
	}{
		{"informer success", namespace,
			nil, []operations{{"abc", "add"}}, map[string]bool{"abc": true}},
		{"informer success & event ordering", namespace,
			nil, []operations{
				{"A", "add"}, {"B", "add"}, {"C", "add"},
				{"A", "delete"}, {"B", "modify"}},
			map[string]bool{"B": true, "C": true}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			eventChan := make(chan watch.Event)
			currNodes := map[string]bool{}
			// TODO CAROLINA -- either finish mocking out the clientSet or scratch it
			mockClientSet := &mockClientSet{watcher: &mockWatcher{c: eventChan}}
			wg.Add(len(tt.operations))

			mockNodeHandler := func(node *k8sV1.Node, toUpdate bool) {
				t.Logf("received node %v", node)
				if node != nil {
					if toUpdate {
						currNodes[node.Name] = true
					} else {
						delete(currNodes, node.Name)
					}
				}
				wg.Done()
			}

			n := newNodeInformer(mockClientSet, mockNodeHandler)
			assert.NotNil(t, n)
			go n.startNodeInformer()
			for _, n := range tt.operations {
				node := &k8sV1.Node{
					ObjectMeta: metaV1.ObjectMeta{
						ResourceVersion: "1",
						Name:            n.name,
					}}
				fmt.Println(node)
				if n.action == "delete" {
					eventChan <- watch.Event{
						Type:   watch.Deleted,
						Object: node,
					}

				} /* else if n.action == "update" {
					eventChan <- watch.Event{
						Type:   watch.Modified,
						Object: node,
					}

				} else if n.action == "add" {
					eventChan <- watch.Event{
						Type:   watch.Added,
						Object: node,
					}

				}
				*/
			}
			// assert.Equal(t, currNodes, tt.output)
			n.stopNodeInformer()
			_, ok := (<-n.stop)
			assert.Equal(t, false, ok, "channel not closed properly.")
		})
	}
}
