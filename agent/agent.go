package agent

import "context"

type DecisionAgent interface {
	Decide(ctx context.Context, snap Snapshot) (Decision, error)
}
