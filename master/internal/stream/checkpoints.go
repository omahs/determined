package stream

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/determined-ai/determined/master/pkg/model"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/uptrace/bun"

	"github.com/determined-ai/determined/master/internal/db"
	"github.com/determined-ai/determined/master/pkg/stream"
	"github.com/determined-ai/determined/proto/pkg/checkpointv1"
)

const (
	// CheckpointsDeleteKey specifies the key for delete metrics.
	CheckpointsDeleteKey = "checkpoints_deleted"
	// CheckpointsUpsertKey specifies the key for upsert metrics.
	CheckpointsUpsertKey = "checkpoint"
)

// CheckpointMsg is a stream.Msg. //use checkpoints_v2
// determined:streamable
type CheckpointMsg struct {
	bun.BaseModel `bun:"table:checkpoints_v2"`
	// immutable attributes
	ID           int
	UUID         string
	TaskID       string
	AllocationID *string
	ReportTime   *timestamp.Timestamp
	State        checkpointv1.State
	Resources    JSONB
	Metadata     JSONB
	Size         int

	// metadata
	Seq int64 `bun:"seq" json:"seq"`

	// permission scope
	WorkspaceID int `json:"-"`

	// TrialID
	TrialID int `json:"-"`

	// ExperimentID
	ExperimentID int `json:"-"`
}

// SeqNum returns the sequence number of a CheckpointMsg.
func (c *CheckpointMsg) SeqNum() int64 {
	return c.Seq
}

// UpsertMsg creates a Checkpoint upserted prepared message.
func (c *CheckpointMsg) UpsertMsg() stream.UpsertMsg {
	return stream.UpsertMsg{
		JSONKey: CheckpointsUpsertKey,
		Msg:     c,
	}
}

// DeleteMsg creates a Checkpoint deleted prepared message.
func (c *CheckpointMsg) DeleteMsg() stream.DeleteMsg {
	deleted := strconv.FormatInt(int64(c.ID), 10)
	return stream.DeleteMsg{
		Key:     CheckpointsDeleteKey,
		Deleted: deleted,
	}
}

// CheckpointSubscriptionSpec is what a user submits to define a checkpoint subscription.
// determined:streamable
type CheckpointSubscriptionSpec struct {
	TrialIDs      []int `json:"trial_ids"`
	ExperimentIDs []int `json:"experiment_ids"`
	Since         int64 `json:"since"`
}

func getCheckpointMsgsWithWorkspaceID(checkpointMsgs []*CheckpointMsg) *bun.SelectQuery {
	q := db.Bun().NewSelect().Model(&checkpointMsgs).
		Column("id").
		Column("uuid").
		Column("task_id").
		Column("allocation_id").
		Column("report_time").
		Column("state").
		Column("resources").
		Column("metadata").
		Column("size").
		Column("trial_id_task_id.trial_id").
		Column("trials.experiment_id").
		Column("projects.workspace_id").
		Join("JOIN trial_id_task_id ON trial_id_task_id.task_id = checkpoint_msg.task_id").
		Join("JOIN trials ON trial_id_task_id.trial_id = trials.id").
		Join("JOIN experiments ON trials.experiment_id = experiments.id").
		Join("JOIN projects ON experiments.project_id = projects.id")
	return q
}

// CheckpointCollectStartupMsgs collects CheckpointMsg's that were missed prior to startup.
func CheckpointCollectStartupMsgs(
	ctx context.Context,
	user model.User,
	known string,
	spec CheckpointSubscriptionSpec,
) (
	[]stream.PreparableMessage, error,
) {
	var out []stream.PreparableMessage

	if len(spec.TrialIDs) == 0 && len(spec.ExperimentIDs) == 0 {
		// empty subscription: everything known should be returned as deleted
		out = append(out, stream.DeleteMsg{
			Key:     CheckpointsDeleteKey,
			Deleted: known,
		})
		return out, nil
	}
	// step 0: get user's permitted access scopes
	accessMap, err := AuthZProvider.Get().GetCheckpointStreamableScopes(ctx, user)
	if err != nil {
		return nil, err
	}
	var accessScopes []model.AccessScopeID
	for id, isPermitted := range accessMap {
		if isPermitted {
			accessScopes = append(accessScopes, id)
		}
	}

	permFilter := func(q *bun.SelectQuery) *bun.SelectQuery {
		if accessMap[model.GlobalAccessScopeID] {
			return q
		}
		return q.Where("workspace_id in (?)", bun.In(accessScopes))
	}

	// step 1: calculate all ids matching this subscription
	q := db.Bun().
		NewSelect().
		Table("checkpoints_v2").
		Column("checkpoints_v2.id").
		Join("JOIN trial_id_task_id ON trial_id_task_id.task_id = checkpoints_v2.task_id").
		Join("JOIN trials ON trial_id_task_id.trial_id = trials.id").
		Join("JOIN experiments e ON trials.experiment_id = e.id").
		Join("JOIN projects p ON e.project_id = p.id").
		OrderExpr("trials.id ASC")
	q = permFilter(q)

	// Ignore tmf.Since, because we want appearances, which might not be have seq > spec.Since.
	ws := stream.WhereSince{Since: 0}
	if len(spec.TrialIDs) > 0 {
		ws.Include("trials.id in (?)", bun.In(spec.TrialIDs))
	}
	if len(spec.ExperimentIDs) > 0 {
		ws.Include("experiment_id in (?)", bun.In(spec.ExperimentIDs))
	}
	q = ws.Apply(q)

	var exist []int64
	err = q.Scan(ctx, &exist)
	if err != nil && errors.Cause(err) != sql.ErrNoRows {
		log.Errorf("error: %v\n", err)
		return nil, err
	}

	// step 2: figure out what was missing and what has appeared
	missing, appeared, err := stream.ProcessKnown(known, exist)
	if err != nil {
		return nil, err
	}

	// step 3: hydrate appeared IDs into full CheckpointMsgs
	var checkpointMsgs []*CheckpointMsg
	if len(appeared) > 0 {
		query := getCheckpointMsgsWithWorkspaceID(checkpointMsgs).
			Where("trial_msg.id in (?)", bun.In(appeared))
		query = permFilter(query)
		err := query.Scan(ctx, &checkpointMsgs)
		if err != nil && errors.Cause(err) != sql.ErrNoRows {
			log.Errorf("error: %v\n", err)
			return nil, err
		}
	}

	// step 4: emit deletions and updates to the client
	out = append(out, stream.DeleteMsg{
		Key:     CheckpointsDeleteKey,
		Deleted: missing,
	})
	for _, msg := range checkpointMsgs {
		out = append(out, stream.UpsertMsg{JSONKey: CheckpointsUpsertKey, Msg: msg})
	}
	return out, nil
}

// CheckpointCollectSubscriptionModMsgs scrapes the database when a
// user submits a new CheckpointSubscriptionSpec for initial matches.
func CheckpointCollectSubscriptionModMsgs(ctx context.Context, addSpec CheckpointSubscriptionSpec) (
	[]interface{}, error,
) {
	if len(addSpec.TrialIDs) == 0 && len(addSpec.ExperimentIDs) == 0 {
		return nil, nil
	}
	var checkpointMsgs []*CheckpointMsg
	q := getCheckpointMsgsWithWorkspaceID(checkpointMsgs)

	// Use WhereSince to build a complex WHERE clause.
	ws := stream.WhereSince{Since: addSpec.Since}
	if len(addSpec.TrialIDs) > 0 {
		ws.Include("id in (?)", bun.In(addSpec.TrialIDs))
	}
	if len(addSpec.ExperimentIDs) > 0 {
		ws.Include("experiment_id in (?)", bun.In(addSpec.ExperimentIDs))
	}
	q = ws.Apply(q)

	err := q.Scan(ctx)
	if err != nil && errors.Cause(err) != sql.ErrNoRows {
		log.Errorf("error: %v\n", err)
		return nil, err
	}

	var out []interface{}
	for _, msg := range checkpointMsgs {
		out = append(out, msg.UpsertMsg())
	}
	return out, nil
}

// CheckpointFilterMaker tracks the Checkpoint and experiment id's that are to be filtered for.
type CheckpointFilterMaker struct {
	TrialIds      map[int]bool
	ExperimentIds map[int]bool
}

// NewCheckpointFilterMaker creates a new FilterMaker.
func NewCheckpointFilterMaker() FilterMaker[*CheckpointMsg, CheckpointSubscriptionSpec] {
	return &CheckpointFilterMaker{make(map[int]bool), make(map[int]bool)}
}

// AddSpec adds CheckpointIds and ExperimentIds specified in CheckpointSubscriptionSpec.
func (ts *CheckpointFilterMaker) AddSpec(spec CheckpointSubscriptionSpec) {
	for _, id := range spec.TrialIDs {
		ts.TrialIds[id] = true
	}
	for _, id := range spec.ExperimentIDs {
		ts.ExperimentIds[id] = true
	}
}

// DropSpec removes CheckpointIds and ExperimentIds specified in CheckpointSubscriptionSpec.
func (ts *CheckpointFilterMaker) DropSpec(spec CheckpointSubscriptionSpec) {
	for _, id := range spec.TrialIDs {
		delete(ts.TrialIds, id)
	}
	for _, id := range spec.ExperimentIDs {
		delete(ts.ExperimentIds, id)
	}
}

// MakeFilter returns a function that determines if a CheckpointMsg based on
// the CheckpointFilterMaker's spec.
func (ts *CheckpointFilterMaker) MakeFilter() func(*CheckpointMsg) bool {
	// Should this filter even run?
	if len(ts.TrialIds) == 0 && len(ts.ExperimentIds) == 0 {
		return nil
	}

	// Make a copy of the maps, because the filter must run safely off-thread.
	checkpointIds := make(map[int]bool)
	experimentIds := make(map[int]bool)
	for id := range ts.TrialIds {
		checkpointIds[id] = true
	}
	for id := range ts.ExperimentIds {
		experimentIds[id] = true
	}

	// return a closure around our copied maps
	return func(msg *CheckpointMsg) bool {
		if _, ok := checkpointIds[msg.ID]; ok {
			return true
		}
		if _, ok := experimentIds[msg.ExperimentID]; ok {
			return true
		}
		return false
	}
}

// CheckpointMakePermissionFilter returns a function that checks if a CheckpointMsg
// is in scope of the user permissions.
func CheckpointMakePermissionFilter(ctx context.Context, user model.User) (func(*CheckpointMsg) bool, error) {
	accessScopeSet, err := AuthZProvider.Get().GetCheckpointStreamableScopes(ctx, user)
	if err != nil {
		return nil, err
	}

	switch {
	case accessScopeSet[model.GlobalAccessScopeID]:
		// user has global access for viewing checkpoints
		return func(msg *CheckpointMsg) bool { return true }, nil
	default:
		return func(msg *CheckpointMsg) bool {
			return accessScopeSet[model.AccessScopeID(msg.WorkspaceID)]
		}, nil
	}
}
