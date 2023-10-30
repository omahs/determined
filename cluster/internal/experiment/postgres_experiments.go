package experiment

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/determined-ai/determined/cluster/internal/api"
	"github.com/determined-ai/determined/cluster/internal/authz"
	"github.com/determined-ai/determined/cluster/internal/db"
	"github.com/determined-ai/determined/cluster/internal/grpcutil"
	"github.com/determined-ai/determined/cluster/pkg/model"
)

// GetExperimentAndCheckCanDoActions fetches an experiment and performs auth checks.
func GetExperimentAndCheckCanDoActions(
	ctx context.Context,
	expID int,
	actions ...func(context.Context, model.User, *model.Experiment) error,
) (*model.Experiment, model.User, error) {
	curUser, _, err := grpcutil.GetUser(ctx)
	if err != nil {
		return nil, model.User{}, err
	}

	e, err := db.ExperimentByID(ctx, expID)
	expNotFound := api.NotFoundErrs("experiment", fmt.Sprint(expID), true)
	if errors.Is(err, db.ErrNotFound) {
		return nil, model.User{}, expNotFound
	} else if err != nil {
		return nil, model.User{}, err
	}

	if err = AuthZProvider.Get().CanGetExperiment(ctx, *curUser, e); err != nil {
		return nil, model.User{}, authz.SubIfUnauthorized(err, expNotFound)
	}

	for _, action := range actions {
		if err = action(ctx, *curUser, e); err != nil {
			return nil, model.User{}, status.Errorf(codes.PermissionDenied, err.Error())
		}
	}
	return e, *curUser, nil
}
