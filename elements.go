package main

import "encoding/json"

// TODO - allow to be number as well
type id string

type Element struct {
	ID    id     `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

type metaData struct {
	*Element
	ProjectRoot string `json:"projectRoot"`
}

type document struct {
	*Element
	URI string `json:"uri"`
}

type documentRange struct {
	*Element
	Start position `json:"start"`
	End   position `json:"end"`
}

type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type edge11 struct {
	*Element
	OutV id `json:"outV"`
	InV  id `json:"inV"`
}

type edge1n struct {
	*Element
	OutV id   `json:"outV"`
	InVs []id `json:"inVs"`
}

type itemEdge struct {
	*Element
	OutV     id   `json:"outV"`
	InVs     []id `json:"inVs"`
	Document id   `json:"document"`
}

func parseElement(line string) (*Element, error) {
	element := &Element{}
	if err := json.Unmarshal([]byte(line), &element); err != nil {
		return nil, err
	}

	return element, nil
}

func parseMetaData(line string) (*metaData, error) {
	metaData := &metaData{}
	if err := json.Unmarshal([]byte(line), &metaData); err != nil {
		return nil, err
	}

	return metaData, nil
}

func parseDocument(line string) (*document, error) {
	document := &document{}
	if err := json.Unmarshal([]byte(line), &document); err != nil {
		return nil, err
	}

	return document, nil
}

func parseDocumentRange(line string) (*documentRange, error) {
	documentRange := &documentRange{}
	if err := json.Unmarshal([]byte(line), &documentRange); err != nil {
		return nil, err
	}

	return documentRange, nil
}

func parseEdge11(line string) (*edge11, error) {
	edge := &edge11{}
	if err := json.Unmarshal([]byte(line), &edge); err != nil {
		return nil, err
	}

	return edge, nil
}

func parseEdge1n(line string) (*edge1n, error) {
	edge := &edge1n{}
	if err := json.Unmarshal([]byte(line), &edge); err != nil {
		return nil, err
	}

	return edge, nil
}

func parseItemEdge(line string) (*itemEdge, error) {
	edge := &itemEdge{}
	if err := json.Unmarshal([]byte(line), &edge); err != nil {
		return nil, err
	}

	return edge, nil
}

func parseEdge(line string) (*edge1n, error) {
	edge1n := &edge1n{}
	if err := json.Unmarshal([]byte(line), &edge1n); err != nil {
		return nil, err
	}

	if len(edge1n.InVs) == 0 {
		edge11 := &edge11{}
		if err := json.Unmarshal([]byte(line), &edge11); err != nil {
			return nil, err
		}

		edge1n.InVs = append(edge1n.InVs, edge11.InV)
	}

	return edge1n, nil
}
