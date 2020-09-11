package main

import (
	"os"

	"github.com/alecthomas/kingpin"
)

var app = kingpin.New(
	"lsif-test",
	"lsif-test is test runner for validating LSIF indexer output.",
).Version(version)

var (
	indexFile *os.File
	testsFile *os.File
)

func init() {
	app.HelpFlag.Short('h')
	app.VersionFlag.Short('v')
	app.HelpFlag.Hidden()

	app.Arg("index-file", "The LSIF index to visualize.").Default("dump.lsif").FileVar(&indexFile)
	app.Arg("tests-file", "The test specification file.").Default("tests.yaml").FileVar(&testsFile)
}

func parseArgs(args []string) (err error) {
	if _, err := app.Parse(args); err != nil {
		return err
	}

	return nil
}
