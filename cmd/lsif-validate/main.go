package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/sourcegraph/lsif-test/assets"
	"github.com/sourcegraph/lsif-test/validation"
	"github.com/xeipuuv/gojsonschema"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}
}

func run() error {
	validator, err := getValidator()
	if err != nil {
		return fmt.Errorf("get validator: %v", err)
	}

	if err := validate(os.Args[1], validator.ValidateLine); err != nil {
		return err
	}

	if err := validator.ValidateGraph(); err != nil {
		return err
	}

	return nil
}

func getValidator() (*validation.Validator, error) {
	schema, err := getSchema()
	if err != nil {
		return nil, fmt.Errorf("get schema: %v", err)
	}

	return validation.NewValidator(schema), nil
}

func getSchema() (*gojsonschema.Schema, error) {
	content, err := assets.Asset("lsif.schema.json")
	if err != nil {
		return nil, fmt.Errorf("asset: %v", err)
	}

	schema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(string(content)))
	if err != nil {
		return nil, fmt.Errorf("new schema: %v", err)
	}

	return schema, nil
}

func validate(filename string, validator validation.LineValidator) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("os open: %v", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := validator(scanner.Text()); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner: %v", err)
	}

	return nil
}
