//go:build integration
// +build integration

package internal

import (
	"testing"

	"github.com/uptrace/bun"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/determined-ai/determined/master/internal/db"
	"github.com/determined-ai/determined/master/internal/mocks"
	"github.com/determined-ai/determined/master/internal/sproto"
	"github.com/determined-ai/determined/proto/pkg/apiv1"
	"github.com/determined-ai/determined/proto/pkg/resourcepoolv1"
)

// var rAuthZ *mocks.ResourceManagerAuthZ.
func TestAuthZGetResourcePools(t *testing.T) {
	return
}

func TestPostBindingFailures(t *testing.T) {
	api, _, ctx := setupAPITest(t, nil)
	var mockRM mocks.ResourceManager
	api.m.rm = &mockRM

	_, err := api.PostWorkspace(ctx, &apiv1.PostWorkspaceRequest{Name: "bindings_test_workspace_1"})
	if err != nil {
		logrus.Error("error posting workspace with name bindings_test_workspace_1 " +
			"(workspace may already exist)")
	}
	_, err = api.PostWorkspace(ctx, &apiv1.PostWorkspaceRequest{Name: "bindings_test_workspace_2"})
	if err != nil {
		logrus.Error("error posting workspace with name bindings_test_workspace_2 " +
			"(workspace my already exist)")
	}

	defer func() {
		workspaces := []string{"bindings_test_workspace_1", "bindings_test_workspace_2"}
		_, err := db.Bun().NewDelete().Table("workspaces").
			Where("name IN (?)", bun.In(workspaces)).
			Exec(ctx)
		if err != nil {
			logrus.Errorf("Error deleting the following workspaces: %s", workspaces)
		}
	}()

	// test resource pools on workspaces that do not exist
	mockRM.On("GetDefaultComputeResourcePool", mock.Anything, mock.Anything).
		Return(sproto.GetDefaultComputeResourcePoolResponse{}, nil).Once()
	mockRM.On("GetDefaultAuxResourcePool", mock.Anything, mock.Anything).
		Return(sproto.GetDefaultAuxResourcePoolResponse{}, nil).Once()
	_, err = api.BindRPToWorkspace(ctx, &apiv1.BindRPToWorkspaceRequest{
		ResourcePoolName: "testRP",
		WorkspaceNames:   []string{"nonexistent_workspace"},
	})
	require.ErrorContains(t, err, "the following workspaces do not exist: [nonexistent_workspace]")

	// test resource pool doesn't exist
	mockRM.On("GetDefaultComputeResourcePool", mock.Anything, mock.Anything).
		Return(sproto.GetDefaultComputeResourcePoolResponse{}, nil).Once()
	mockRM.On("GetDefaultAuxResourcePool", mock.Anything, mock.Anything).
		Return(sproto.GetDefaultAuxResourcePoolResponse{}, nil).Twice()
	mockRM.On("GetResourcePools", mock.Anything, mock.Anything).
		Return(&apiv1.GetResourcePoolsResponse{
			ResourcePools: []*resourcepoolv1.ResourcePool{},
		}, nil).Once()
	_, err = api.BindRPToWorkspace(ctx, &apiv1.BindRPToWorkspaceRequest{
		ResourcePoolName: "testRP",
		WorkspaceNames:   []string{"bindings_test_workspace_1", "bindings_test_workspace_2"},
	})

	require.ErrorContains(t, err, "pool with name testRP doesn't exist in config")

	// test resource pool is a default resource pool
	mockRM.On("GetDefaultComputeResourcePool", mock.Anything, mock.Anything).
		Return(sproto.GetDefaultComputeResourcePoolResponse{PoolName: "testRP"}, nil).Twice()
	mockRM.On("GetResourcePools", mock.Anything, mock.Anything).
		Return(&apiv1.GetResourcePoolsResponse{
			ResourcePools: []*resourcepoolv1.ResourcePool{{Name: "testRP"}},
		}, nil).Twice()

	_, err = api.BindRPToWorkspace(ctx, &apiv1.BindRPToWorkspaceRequest{
		ResourcePoolName: "testRP",
		WorkspaceNames:   []string{"bindings_test_workspace_1", "bindings_test_workspace_2"},
	})

	require.ErrorContains(t, err, "default resource pool testRP cannot be bound to any workspace")

	mockRM.On("GetDefaultAuxResourcePool", mock.Anything, mock.Anything).
		Return(sproto.GetDefaultAuxResourcePoolResponse{PoolName: "testRP"}, nil).Once()

	_, err = api.BindRPToWorkspace(ctx, &apiv1.BindRPToWorkspaceRequest{
		ResourcePoolName: "testRP",
		WorkspaceNames:   []string{"bindings_test_workspace_1", "bindings_test_workspace_2"},
	})

	require.ErrorContains(t, err, "default resource pool testRP cannot be bound to any workspace")

	// test no resource pool specified
	mockRM.On("GetDefaultComputeResourcePool", mock.Anything, mock.Anything).
		Return(sproto.GetDefaultComputeResourcePoolResponse{PoolName: "testRP"}, nil).Once()
	mockRM.On("GetDefaultAuxResourcePool", mock.Anything, mock.Anything).
		Return(sproto.GetDefaultAuxResourcePoolResponse{PoolName: "testRP"}, nil).Once()
	mockRM.On("GetResourcePools", mock.Anything, mock.Anything).
		Return(&apiv1.GetResourcePoolsResponse{
			ResourcePools: []*resourcepoolv1.ResourcePool{{Name: "testRP"}},
		}, nil).Once()
	_, err = api.BindRPToWorkspace(ctx, &apiv1.BindRPToWorkspaceRequest{
		ResourcePoolName: "",
		WorkspaceNames:   []string{"bindings_test_workspace_1", "bindings_test_workspace_2"},
	})

	require.ErrorContains(t, err, "doesn't exist in config")

	if err != nil {
		return
	}
	return
}

func TestPostBindingSucceeds(t *testing.T) {
	api, _, ctx := setupAPITest(t, nil)
	var mockRM mocks.ResourceManager
	api.m.rm = &mockRM

	_, err := api.PostWorkspace(ctx, &apiv1.PostWorkspaceRequest{Name: "bindings_test_workspace_1"})
	if err != nil {
		logrus.Error("error posting workspace with name bindings_test_workspace_1 " +
			"(workspace may already exist)")
	}
	_, err = api.PostWorkspace(ctx, &apiv1.PostWorkspaceRequest{Name: "bindings_test_workspace_2"})
	if err != nil {
		logrus.Error("error posting workspace with name bindings_test_workspace_2 " +
			"(workspace my already exist)")
	}

	defer func() {
		workspaces := []string{"bindings_test_workspace_1", "bindings_test_workspace_2"}
		_, err := db.Bun().NewDelete().Table("workspaces").
			Where("name IN (?)", bun.In(workspaces)).
			Exec(ctx)
		if err != nil {
			logrus.Errorf("Error deleting the following workspaces: %s", workspaces)
		}
	}()

	// test resource pool is a default resource pool
	mockRM.On("GetDefaultComputeResourcePool", mock.Anything, mock.Anything).
		Return(sproto.GetDefaultComputeResourcePoolResponse{}, nil).Twice()
	mockRM.On("GetDefaultAuxResourcePool", mock.Anything, mock.Anything).
		Return(sproto.GetDefaultAuxResourcePoolResponse{}, nil).Twice()
	mockRM.On("GetResourcePools", mock.Anything, mock.Anything).
		Return(&apiv1.GetResourcePoolsResponse{
			ResourcePools: []*resourcepoolv1.ResourcePool{{Name: "testRP"}},
		}, nil).Twice()

	_, err = api.BindRPToWorkspace(ctx, &apiv1.BindRPToWorkspaceRequest{
		ResourcePoolName: "testRP",
		WorkspaceNames:   []string{"bindings_test_workspace_1"},
	})
	require.NoError(t, err)

	// test no workspaces specified
	mockRM.On("GetDefaultAuxResourcePool", mock.Anything, mock.Anything).
		Return(sproto.GetDefaultAuxResourcePoolResponse{PoolName: "testRP"}, nil).Once()

	_, err = api.BindRPToWorkspace(ctx, &apiv1.BindRPToWorkspaceRequest{
		ResourcePoolName: "testRP",
		WorkspaceNames:   []string{},
	})

	require.NoError(t, err)

	if err != nil {
		return
	}
	return
}

func TestGetBindings(t *testing.T) {
	return
}

func TestPatchBindings(t *testing.T) {
	return
}

func TestDeleteBindings(t *testing.T) {
	return
}
