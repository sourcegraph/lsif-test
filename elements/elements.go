package elements

import "encoding/json"

// TODO - allow to be number as well
type ID string

type Element struct {
	ID    ID     `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

type MetaData struct {
	*Element
	ProjectRoot string `json:"projectRoot"`
}

type Document struct {
	*Element
	URI string `json:"uri"`
}

type DocumentRange struct {
	*Element
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Edge11 struct {
	*Element
	OutV ID `json:"outV"`
	InV  ID `json:"inV"`
}

type Edge1n struct {
	*Element
	OutV ID   `json:"outV"`
	InVs []ID `json:"inVs"`
}

type ItemEdge struct {
	*Element
	OutV     ID   `json:"outV"`
	InVs     []ID `json:"inVs"`
	Document ID   `json:"document"`
}

func ParseElement(line string) (*Element, error) {
	element := &Element{}
	if err := json.Unmarshal([]byte(line), &element); err != nil {
		return nil, err
	}

	return element, nil
}

func ParseMetaData(line string) (*MetaData, error) {
	metaData := &MetaData{}
	if err := json.Unmarshal([]byte(line), &metaData); err != nil {
		return nil, err
	}

	return metaData, nil
}

func ParseDocument(line string) (*Document, error) {
	document := &Document{}
	if err := json.Unmarshal([]byte(line), &document); err != nil {
		return nil, err
	}

	return document, nil
}

func ParseDocumentRange(line string) (*DocumentRange, error) {
	documentRange := &DocumentRange{}
	if err := json.Unmarshal([]byte(line), &documentRange); err != nil {
		return nil, err
	}

	return documentRange, nil
}

func ParseEdge11(line string) (*Edge11, error) {
	edge := &Edge11{}
	if err := json.Unmarshal([]byte(line), &edge); err != nil {
		return nil, err
	}

	return edge, nil
}

func ParseEdge1n(line string) (*Edge1n, error) {
	edge := &Edge1n{}
	if err := json.Unmarshal([]byte(line), &edge); err != nil {
		return nil, err
	}

	return edge, nil
}

func ParseItemEdge(line string) (*ItemEdge, error) {
	edge := &ItemEdge{}
	if err := json.Unmarshal([]byte(line), &edge); err != nil {
		return nil, err
	}

	return edge, nil
}

func ParseEdge(line string) (*Edge1n, error) {
	edge1n := &Edge1n{}
	if err := json.Unmarshal([]byte(line), &edge1n); err != nil {
		return nil, err
	}

	if len(edge1n.InVs) == 0 {
		edge11 := &Edge11{}
		if err := json.Unmarshal([]byte(line), &edge11); err != nil {
			return nil, err
		}

		edge1n.InVs = append(edge1n.InVs, edge11.InV)
	}

	return edge1n, nil
}
