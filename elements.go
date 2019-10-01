package main

import "encoding/json"

// TODO - allow to be number as well
type id string

type element struct {
	ID    id     `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

type metaData struct {
	ProjectRoot string `json:"projectRoot"`
}

type document struct {
	URI string `json:"uri"`
}

type documentRange struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type edge11 struct {
	OutV id `json:"outV"`
	InV  id `json:"inV"`
}

type edge1n struct {
	OutV id   `json:"outV"`
	InVs []id `json:"inVs"`
}

func parseElement(line string) (*element, error) {
	element := &element{}
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
