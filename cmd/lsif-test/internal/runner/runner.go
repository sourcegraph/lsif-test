package runner

import (
	"fmt"
	"io"
	"os"

	reader "github.com/sourcegraph/lsif-protocol/reader"
	"github.com/sourcegraph/lsif-test/cmd/lsif-test/internal/search"
	reader2 "github.com/sourcegraph/lsif-test/internal/reader"
)

type Runner struct {
	Context *RunnerContext
}

func (v *Runner) Run(indexFile, testsFile io.Reader) error {
	testSpecs, err := ReadTestSpecs(testsFile)
	if err != nil {
		return err
	}

	if err := reader2.Read(indexFile, v.Context.Stasher, nil, nil); err != nil {
		return err
	}

	var vertices []reader.Element
	_ = v.Context.Stasher.Vertices(func(lineContext reader2.LineContext) bool {
		vertices = append(vertices, lineContext.Element)
		return true
	})

	var edges []reader.Element
	_ = v.Context.Stasher.Edges(func(lineContext reader2.LineContext, edge reader.Edge) bool {
		edges = append(edges, lineContext.Element)
		return true
	})

	for _, testSpec := range testSpecs {
		rangeData := search.GatherRangeDataFromPosition(vertices, edges, testSpec.Path, testSpec.Line, testSpec.Character)
		if len(rangeData) != 1 {
			fmt.Printf("error: no or overlapping range data\n")
			os.Exit(1)
		}

		if rangeData[0].HoverText != testSpec.HoverText {
			fmt.Printf("error: bad hover text\n")
			os.Exit(1)
		}

		// TODO - test definitions
		// TODO - test references
		// TODO - test monikers
	}

	return nil
}
