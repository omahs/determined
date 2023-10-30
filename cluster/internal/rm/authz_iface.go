package rm

import (
	"context"

	"github.com/determined-ai/determined/cluster/internal/authz"
	"github.com/determined-ai/determined/cluster/pkg/generatedproto/resourcepoolv1"
	"github.com/determined-ai/determined/cluster/pkg/model"
)

// ResourceManagerAuthZ is the interface for resource manager authorization.
type ResourceManagerAuthZ interface {
	// GET /api/v1/resource-pools
	FilterResourcePools(
		ctx context.Context, curUser model.User, resourcePools []*resourcepoolv1.ResourcePool,
		accessibleWorkspaces []int32,
	) ([]*resourcepoolv1.ResourcePool, error)
}

// AuthZProvider provides ResourceManagerAuthZ implementations.
var AuthZProvider authz.AuthZProviderType[ResourceManagerAuthZ]
