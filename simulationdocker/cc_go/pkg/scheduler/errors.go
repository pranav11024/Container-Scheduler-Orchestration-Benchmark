// pkg/scheduler/errors.go - Scheduler error definitions
package scheduler

import "errors"

var (
	ErrNoSuitableNode = errors.New("no suitable node found")
)
