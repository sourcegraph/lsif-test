package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/sourcegraph/lsif-test/elements"
)

const version = "0.1.0"

func main() {
	if err := realMain(); err != nil {
		fmt.Fprint(os.Stderr, fmt.Sprintf("error: %v\n", err))
		os.Exit(1)
	}
}

func realMain() error {
	app := kingpin.New("lsif-extractor", "lsif-extractor extracts a subgraph containing a given id.").Version(version)
	id := app.Arg("id", "The identifier to extract.").Required().String()
	dumpFile := app.Arg("dump-file", "The LSIF output to read.").Default("dump.lsif").File()
	depth := app.Arg("diameter", "The maximum shortest path in the output graph.").Default("5").Int()
	bufferCapacity := app.Flag("buffer-capacity", "Set the max line size.").Default("1000000").Int()

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	defer (*dumpFile).Close()

	scanner := bufio.NewScanner(*dumpFile)
	scanner.Buffer(make([]byte, *bufferCapacity), *bufferCapacity)
	graph, err := buildGraph(scanner)
	if err != nil {
		return err
	}

	display(extractSubgraph(graph, elements.ID{*id}, *depth))
	return nil
}

type Graph struct {
	vertices     map[elements.ID]string
	edges        map[elements.ID][]elements.ID
	reverseEdges map[elements.ID][]elements.ID
}

func buildGraph(scanner *bufio.Scanner) (Graph, error) {
	vertices := map[elements.ID]string{}
	edges := map[elements.ID][]elements.ID{}
	reverseEdges := map[elements.ID][]elements.ID{}

	for scanner.Scan() {
		line := scanner.Text()

		element, err := elements.ParseElement(line)
		if err != nil {
			return Graph{}, err
		}

		if element.Type == "vertex" {
			vertices[element.ID] = element.Label
		} else {
			edge, err := elements.ParseEdge(line)
			if err != nil {
				return Graph{}, err
			}

			if edge.Label != "contains" {
				edges[edge.OutV] = append(edges[edge.OutV], edge.InVs...)
				for _, inV := range edge.InVs {
					reverseEdges[inV] = append(reverseEdges[inV], edge.OutV)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return Graph{}, fmt.Errorf("scanner: %v", err)
	}

	return Graph{vertices, edges, reverseEdges}, nil
}

func extractSubgraph(g Graph, id elements.ID, depth int) Graph {
	var add func(elements.ID, int)
	vertices := map[elements.ID]string{}

	add = func(id elements.ID, remaining int) {
		if _, ok := vertices[id]; ok || remaining == 0 {
			return
		}

		vertices[id] = g.vertices[id]

		for _, n := range g.edges[id] {
			add(n, remaining-1)
		}

		for _, n := range g.reverseEdges[id] {
			add(n, remaining-1)
		}
	}
	add(id, depth)

	edges := map[elements.ID][]elements.ID{}
	for id := range vertices {
		for _, inV := range g.edges[id] {
			if _, ok := vertices[inV]; ok {
				edges[id] = append(edges[id], inV)
			}
		}
	}

	return Graph{vertices, edges, nil}
}

func display(g Graph) {
	fmt.Printf("digraph {\n")

	for id, label := range g.vertices {
		fmt.Printf("\t"+`%s [label="%s\n%s"]`+"\n", id, label, id)
	}

	fmt.Printf("\n")

	for outV, inVs := range g.edges {
		for _, inV := range inVs {
			fmt.Printf("\t"+"%s -> %s"+"\n", outV, inV)
		}
	}

	fmt.Printf("}\n")
}
