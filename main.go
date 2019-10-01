package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/xeipuuv/gojsonschema"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}
}

func run() error {
	// TODO - use go-bindata
	schema, err := gojsonschema.NewSchema(gojsonschema.NewReferenceLoader("file://./lsif.schema.json"))
	if err != nil {
		return err
	}

	validator := NewValidator(schema)

	if err := validate(os.Args[1], validator.ValidateLine); err != nil {
		return err
	}

	if err := validator.Process(); err != nil {
		return err
	}

	return nil
}

func validate(filename string, validator lineValidator) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := validator(scanner.Text()); err != nil {
			return err
		}
	}

	return scanner.Err()
}
