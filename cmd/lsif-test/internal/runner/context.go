package runner

import (
	"github.com/sourcegraph/lsif-test/internal/reader"
)

type RunnerContext struct {
	Stasher *reader.Stasher
}

func NewRunnerContext() *RunnerContext {
	return &RunnerContext{
		Stasher: reader.NewStasher(),
	}
}
