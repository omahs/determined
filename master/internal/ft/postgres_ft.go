package ft

import (
	"github.com/determined-ai/determined/proto/pkg/agentv1"
)

// we could probably implement just in memory for phase one.

// AddAlert persists an alert to the database.
func AddAlert(alert *agentv1.RunAlert) (err error) {
	return
}

// GetAlerts retrieves all alerts for a given job.
func GetAlerts(jobID string) (alerts []*agentv1.RunAlert, err error) {
	// has allocation id and task id, rp, node id
	return
}
