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
	disableJSONSchema := app.Flag("disable-jsonschema", "Turn off JSON schema validation").Bool()
	stopOnError := app.Flag("stop-on-error", "Stop validation after the first error.").Bool()

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	defer (*dumpFile).Close()

	schema, err := getSchema()
	if err != nil {
		return fmt.Errorf("schema: %v", err)
	}

	valid := true
	scanner := bufio.NewScanner(*dumpFile)
	validator := validation.NewValidator(schema, *disableJSONSchema)

	for scanner.Scan() {
		if !validator.ValidateLine(scanner.Text()) {
			valid = false

			if *stopOnError {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner: %v", err)
	}

	if valid {
		if !validator.ValidateGraph(*stopOnError) {
			valid = false
		}
	}

	numVertices, numEdges := validator.Stats()
	fmt.Printf("Processed %d and %d edges\n", numVertices, numEdges)

	errors := validator.Errors()
	fmt.Printf("Found %d errors\n\n", len(errors))

	for i, err := range errors {
		fmt.Printf("%d) %s\n", i+1, err.Message)
		for _, lineContext := range err.RelevantLines {
			fmt.Printf("\ton line #%d: %s\n", lineContext.LineIndex, lineContext.LineText)
		}
	}

	if len(errors) > 0 {
		fmt.Printf("\n")
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
