package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/efritz/pentimento"
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
	app := kingpin.New("lsif-validate", "lsif-validate is validator for LSIF indexer output.").Version(version)
	dumpFile := app.Arg("dump-file", "The LSIF output to validate.").Default("dump.lsif").File()
	disableJSONSchema := app.Flag("disable-jsonschema", "Turn off JSON schema validation").Bool()
	stopOnError := app.Flag("stop-on-error", "Stop validation after the first error.").Bool()
	bufferCapacity := app.Flag("buffer-capacity", "Set the max line size.").Default("1000000").Int()

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	defer (*dumpFile).Close()

	header := fmt.Sprintf("Validating LSIF dump at %s:", (*dumpFile).Name())

	schema, err := getSchema()
	if err != nil {
		return fmt.Errorf("schema: %v", err)
	}

	valid := true
	scanner := bufio.NewScanner(*dumpFile)
	scanner.Buffer(make([]byte, *bufferCapacity), *bufferCapacity)
	validator := validation.NewValidator(schema, *disableJSONSchema)

	formatStats := func() string {
		numVertices, numEdges, numErrors := validator.Stats()
		return fmt.Sprintf(
			"Processed %d lines, %d vertices and %d edges.\nFound %d errors.\n\n",
			numVertices+numEdges,
			numVertices,
			numEdges,
			numErrors,
		)
	}

	withProgressUpdate := func(status string, f func()) {
		pentimento.PrintProgress(func(p *pentimento.Printer) error {
			done := make(chan struct{})
			go func() {
				defer close(done)
				f()
			}()

		loop:
			for {
				content := pentimento.NewContent()
				content.AddLine("%s %s%s", header, status, pentimento.Dots)
				content.AddLine(formatStats())
				p.WriteContent(content)

				select {
				case <-done:
					break loop
				case <-time.After(time.Second / 4):
				}
			}

			p.Reset()
			return nil
		})
	}

	withProgressUpdate("processing individual lines", func() {
		for scanner.Scan() {
			if !validator.ValidateLine(scanner.Text()) {
				valid = false
				if *stopOnError {
					break
				}
			}
		}
	})

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner: %v", err)
	}

	completedValidation := false
	if valid {
		withProgressUpdate("processing relationships", func() {
			if !validator.ValidateGraph(*stopOnError) {
				valid = false
			}
		})
		completedValidation = true
	}

	fmt.Printf("%s done.\n", header)
	fmt.Printf(formatStats())

	hasErrors := false
	for i, err := range validator.Errors() {
		hasErrors = true
		fmt.Printf("%d) %s\n", i+1, err.Message)
		for _, lineContext := range err.RelevantLines {
			fmt.Printf("\ton line #%d: %s\n", lineContext.LineIndex, lineContext.LineText)
		}
	}

	if !completedValidation {
		fmt.Printf("\n")
		fmt.Println("WARNING Partial validation! Fix errors and re-run to continue validation.")
	}

	if hasErrors {
		fmt.Printf("\n")
		os.Exit(1)
	}

	fmt.Printf(":)\n")
	return nil
}

func getSchema() (*gojsonschema.Schema, error) {
	content, err := assets.Asset("lsif.schema.json")
	if err != nil {
		return nil, err
	}

	return gojsonschema.NewSchema(gojsonschema.NewStringLoader(string(content)))
}
