package search

import (
	"strings"

	"github.com/sourcegraph/lsif-protocol/reader"
)

type RangeData struct {
	HoverText   string
	Definitions []Location
	References  []Location
}

type Location struct {
	Path           string
	StartLine      int
	StartCharacter int
	EndLine        int
	EndCharacter   int
}

func GatherRangeDataFromPosition(vertices, edges []reader.Element, path string, line, character int) []RangeData {
	var rangeData []RangeData
	for _, rangeID := range findRangeByPosition(vertices, edges, path, line, character) {
		rangeData = append(rangeData, gatherRangeDataFromID(vertices, edges, rangeID))
	}

	return rangeData
}

func gatherRangeDataFromID(vertices, edges []reader.Element, rangeID int) RangeData {
	return RangeData{
		HoverText:   findHoverTextByRangeID(vertices, edges, rangeID),
		Definitions: findDefinitionLocationsByRangeID(vertices, edges, rangeID),
		References:  findReferenceLocationsByRangeID(vertices, edges, rangeID),
	}
}

func findRangeByPosition(vertices, edges []reader.Element, path string, line, character int) []int {
	projectRoot := getProjectRoot(vertices, edges)

	for _, vertex := range vertices {
		if vertex.Label == "document" && strings.TrimPrefix(vertex.Payload.(string), projectRoot) == path {
			return findOverlappingRanges(vertices, edges, findRangesContainedByDocument(vertices, edges, vertex.ID), line, character)
		}
	}

	return nil
}

func getProjectRoot(vertices, edges []reader.Element) string {
	for _, vertex := range vertices {
		if vertex.Label == "metaData" {
			projectRoot := vertex.Payload.(reader.MetaData).ProjectRoot
			if !strings.HasSuffix(projectRoot, "/") {
				projectRoot += "/"
			}

			return projectRoot
		}
	}

	return ""
}

func findRangesContainedByDocument(vertices, edges []reader.Element, documentID int) []int {
	for _, edge := range edges {
		if payload := edge.Payload.(reader.Edge); edge.Label == "contains" && payload.OutV == documentID {
			return payload.InVs
		}
	}

	return nil
}

func findOverlappingRanges(vertices, edges []reader.Element, rangeIDs []int, line, character int) []int {
	rangeIDMap := map[int]struct{}{}
	for _, rangeID := range rangeIDs {
		rangeIDMap[rangeID] = struct{}{}
	}

	var intersectingRangeIDs []int
	for _, vertex := range vertices {
		if _, ok := rangeIDMap[vertex.ID]; ok && contains(vertex.Payload.(reader.Range), line, character) {
			intersectingRangeIDs = append(intersectingRangeIDs, vertex.ID)
		}
	}

	return intersectingRangeIDs
}

func contains(r reader.Range, line, character int) bool {
	if r.StartLine > line || r.EndLine < line {
		return false
	}
	if r.StartLine == line && r.StartCharacter > character {
		return false
	}
	if r.EndLine == line && r.EndCharacter < character {
		return false
	}

	return true
}

func findHoverTextByRangeID(vertices, edges []reader.Element, rangeID int) string {
	for _, edge := range edges {
		if payload := edge.Payload.(reader.Edge); edge.Label == "textDocument/hover" && payload.OutV == rangeID {
			return findHoverTextByID(vertices, edges, payload.InV)
		}
	}

	for _, edge := range edges {
		if payload := edge.Payload.(reader.Edge); edge.Label == "next" && payload.OutV == rangeID {
			if hoverText := findHoverTextByRangeID(vertices, edges, payload.InV); hoverText != "" {
				return hoverText
			}
		}
	}

	return ""
}

func findHoverTextByID(vertices, edges []reader.Element, hoverResultID int) string {
	for _, vertex := range vertices {
		if vertex.Label == "hoverResult" && vertex.ID == hoverResultID {
			return vertex.Payload.(string)
		}
	}

	return ""
}

//
//
// TODO
//
//

func findDefinitionLocationsByRangeID(vertices, edges []reader.Element, rangeID int) []Location {
	for _, edge := range edges {
		if payload := edge.Payload.(reader.Edge); edge.Label == "textDocument/definition" && payload.OutV == rangeID {
			return findDefinitionLocationsByID(vertices, edges, payload.InV)
		}
	}

	for _, edge := range edges {
		if payload := edge.Payload.(reader.Edge); edge.Label == "next" && payload.OutV == rangeID {
			if locations := findDefinitionLocationsByRangeID(vertices, edges, payload.InV); len(locations) > 0 {
				return locations
			}
		}
	}

	return nil
}

func findDefinitionLocationsByID(vertices, edges []reader.Element, definitionResultID int) []Location {
	var locations []Location
	for _, edge := range edges {
		if payload := edge.Payload.(reader.Edge); edge.Label == "item" && payload.OutV == definitionResultID {
			for _, inV := range payload.InVs {
				locations = append(locations, findLocationByID(vertices, edges, inV))
			}
		}
	}

	return locations
}

func findReferenceLocationsByRangeID(vertices, edges []reader.Element, rangeID int) []Location {
	for _, edge := range edges {
		if payload := edge.Payload.(reader.Edge); edge.Label == "textDocument/references" && payload.OutV == rangeID {
			return findReferenceLocationsByID(vertices, edges, payload.InV)
		}
	}

	for _, edge := range edges {
		if payload := edge.Payload.(reader.Edge); edge.Label == "next" && payload.OutV == rangeID {
			if locations := findReferenceLocationsByRangeID(vertices, edges, payload.InV); len(locations) > 0 {
				return locations
			}
		}
	}

	return nil
}

func findReferenceLocationsByID(vertices, edges []reader.Element, referenceResultID int) []Location {
	var locations []Location
	for _, edge := range edges {
		if payload := edge.Payload.(reader.Edge); edge.Label == "item" && payload.OutV == referenceResultID {
			for _, inV := range payload.InVs {
				locations = append(locations, findLocationByID(vertices, edges, inV))
			}
		}
	}

	return locations
}

func findLocationByID(vertices, edges []reader.Element, rangeID int) Location {
	for _, vertex := range vertices {
		if vertex.Label == "range" && vertex.ID == rangeID {
			r := vertex.Payload.(reader.Range)

			return Location{
				Path:           findURIContaining(vertices, edges, rangeID),
				StartLine:      r.StartLine,
				StartCharacter: r.StartCharacter,
				EndLine:        r.EndLine,
				EndCharacter:   r.EndCharacter,
			}
		}
	}

	return Location{}
}

func findURIContaining(vertices, edges []reader.Element, rangeID int) string {
	for _, edge := range edges {
		if payload := edge.Payload.(reader.Edge); edge.Label == "contains" {
			for _, inV := range payload.InVs {
				if inV == rangeID {
					projectRoot := getProjectRoot(vertices, edges)

					for _, vertex := range vertices {
						if vertex.Label == "document" && vertex.ID == payload.OutV {
							return strings.TrimPrefix(vertex.Payload.(string), projectRoot)
						}
					}

					return ""
				}
			}
		}
	}

	return ""
}
