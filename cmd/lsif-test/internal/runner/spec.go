package runner

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type TestSpec struct {
	Path        string     `yaml:"path"`
	Line        int        `yaml:"line"`
	Character   int        `yaml:"character"`
	HoverText   string     `yaml:"hoverText"`
	Definitions []Location `yaml:"definitions"`
	References  []Location `yaml:"references"`
}

type Location struct {
	Path           string `yaml:"path"`
	StartLine      int    `yaml:"startLine"`
	StartCharacter int    `yaml:"startCharacter"`
	EndLine        int    `yaml:"endLine"`
	EndCharacter   int    `yaml:"endCharacter"`
}

func ReadTestSpecs(r io.Reader) ([]TestSpec, error) {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var testSpecs []TestSpec
	if err := yaml.Unmarshal(content, &testSpecs); err != nil {
		return nil, err
	}

	return testSpecs, nil
}
