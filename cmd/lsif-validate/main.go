package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/sourcegraph/lsif-test/assets"
	"github.com/sourcegraph/lsif-test/validation"
	"github.com/xeipuuv/gojsonschema"
)

const version = "0.1.0"

func main() {
	if err := realMain(); err != nil {
		fmt.Fprint(os.Stderr, fmt.Sprintf("error: %v\n", err))
		os.Exit(1)
	}
}

func realMain() error {
	app := kingpin.New("lsif-go", "lsif-validate is validator for LSIF indexer output.").Version(version)
	dumpFile := app.Arg("dump-file", "The LSIf output to validate.").Default("data.lsif").File()

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	defer (*dumpFile).Close()

	schema, err := getSchema()
	if err != nil {
		return fmt.Errorf("schema: %v", err)
	}

	validator := validation.NewValidator(schema)

	scanner := bufio.NewScanner(*dumpFile)
	for scanner.Scan() {
		if err := validator.ValidateLine(scanner.Text()); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner: %v", err)
	}

	if err := validator.ValidateGraph(); err != nil {
		return err
	}

	return nil
}

func getSchema() (*gojsonschema.Schema, error) {
	content, err := assets.Asset("lsif.schema.json")
	if err != nil {
		return nil, err
	}

	return gojsonschema.NewSchema(gojsonschema.NewStringLoader(string(content)))
}
