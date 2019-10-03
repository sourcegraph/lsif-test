package elements

import "encoding/json"

type DocumentRange struct {
	*Element
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

func ParseDocumentRange(line string) (*DocumentRange, error) {
	documentRange := &DocumentRange{}
	if err := json.Unmarshal([]byte(line), &documentRange); err != nil {
		return nil, err
	}

	return documentRange, nil
}
