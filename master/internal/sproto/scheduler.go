package sproto

type Blacklist []string

// FittingRequirements allow tasks to specify requirements for their placement.
type FittingRequirements struct {
	// SingleAgent specifies that the task must be located within a single agent.
	SingleAgent bool
	// Blacklist specifies a list of slots to ignore. Device
	Blacklist Blacklist
}
