package main

import (
	"os"

	"github.com/sourcegraph/lsif-test/cmd/lsif-test/internal/runner"
)

func test(indexFile, testsFile *os.File) error {
	ctx := runner.NewRunnerContext()
	runner := &runner.Runner{Context: ctx}
	return runner.Run(indexFile, testsFile)
}
