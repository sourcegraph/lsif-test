package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/fatih/color"
	_ "github.com/lib/pq"
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
	app := kingpin.New("lsif-annotate", "lsif-annotate is an annotator for debugging LSIF indexer output.").Version(version)
	dumpFile := app.Arg("dump-file", "The LSIF output to validate.").Default("dump.lsif").File()
	bufferCapacity := app.Flag("buffer-capacity", "Set the max line size.").Default("1000000").Int()
	docURIToAnnotate := app.Flag("docURI", "The document URI to annotate.").Required().String()

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	defer (*dumpFile).Close()

	db, err := loadDump(*dumpFile, *bufferCapacity)
	if err != nil {
		return err
	}

	results, err := queryAllRaw(db, `
SELECT ranges.data FROM lsif_annotate doc
JOIN lsif_annotate contains ON
	contains.data->>'outV' = doc.data->>'id'
JOIN lsif_annotate ranges ON
	contains.data->'inVs' ? (ranges.data->>'id'::TEXT)
WHERE doc.data->>'uri' = $1
ORDER BY ranges.data->'start'->'line';
`, *docURIToAnnotate)
	if err != nil {
		return err
	}
	rangesByLine := make(map[int][]elements.DocumentRange)
	for _, v := range results {
		var ranje elements.DocumentRange
		json.Unmarshal(v, &ranje)
		if ranges, ok := rangesByLine[ranje.Start.Line]; ok {
			rangesByLine[ranje.Start.Line] = append(ranges, ranje)
		} else {
			rangesByLine[ranje.Start.Line] = []elements.DocumentRange{ranje}
		}
	}

	results, err = queryAllRaw(db, `
SELECT ranges.data FROM lsif_annotate defs
JOIN lsif_annotate items ON
	defs.data->>'id' = items.data->>'outV'
JOIN lsif_annotate ranges ON
	items.data->'inVs' ? (ranges.data->>'id'::TEXT)
WHERE defs.data->>'label' = 'definitionResult'
ORDER BY ranges.data->'start'->'line';
`)
	if err != nil {
		return err
	}
	defRangeIDs := make(map[elements.ID]interface{}, 0)
	for _, v := range results {
		var ranje elements.DocumentRange
		json.Unmarshal(v, &ranje)
		defRangeIDs[ranje.ID] = nil
	}

	results, err = queryAllRaw(db, `
SELECT ranges.data FROM lsif_annotate refs
JOIN lsif_annotate items ON
	refs.data->>'id' = items.data->>'outV'
JOIN lsif_annotate ranges ON
	items.data->'inVs' ? (ranges.data->>'id'::TEXT)
WHERE refs.data->>'label' = 'referenceResult'
ORDER BY ranges.data->'start'->'line';
`)
	if err != nil {
		return err
	}
	refRangeIDs := make(map[elements.ID]interface{}, 0)
	for _, v := range results {
		var ranje elements.DocumentRange
		json.Unmarshal(v, &ranje)
		refRangeIDs[ranje.ID] = nil
	}

	filepath := strings.TrimPrefix(*docURIToAnnotate, "file://")
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	content := string(bytes)
	lines := strings.Split(content, "\n")
	for linenumber, line := range lines {
		padding := strconv.Itoa(len(strconv.Itoa(len(lines))))
		linePrefix := fmt.Sprintf("Line %-"+padding+"d|", linenumber)
		fmt.Println(color.New(color.FgYellow).Sprint(linePrefix) + color.New(color.FgBlack).Sprint(line))

		ranges := rangesByLine[linenumber]
		sort.SliceStable(ranges, func(i, j int) bool {
			return ranges[i].Start.Character > ranges[j].Start.Character
		})

		for _, rainge := range ranges {
			whitespace := strings.Repeat(" ", len(linePrefix)+rainge.Start.Character)
			indicator := color.New(color.FgGreen).Sprint(strings.Repeat("^", rainge.End.Character-rainge.Start.Character))
			label := ""
			label += color.New(color.FgBlue).Sprintf("range id %s character %d", rainge.ID.Value, rainge.Start.Character)
			if _, ok := defRangeIDs[rainge.ID]; ok {
				label += " " + color.New(color.FgHiRed).Sprint("def")
			}
			if _, ok := refRangeIDs[rainge.ID]; ok {
				label += " " + color.New(color.FgHiMagenta).Sprint("ref")
			}
			fmt.Println(whitespace + indicator + " " + label)
		}
	}

	return nil
}

func queryAllRaw(db *sql.DB, query string, args ...interface{}) ([][]byte, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([][]byte, 0)
	for rows.Next() {
		var result []byte
		err := rows.Scan(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func queryAll(db *sql.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rawResults, err := queryAllRaw(db, query, args...)
	if err != nil {
		return nil, err
	}
	results := make([]map[string]interface{}, 0)
	for _, rawResult := range rawResults {
		var result map[string]interface{}
		err = json.Unmarshal(rawResult, &result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func loadDump(dumpFile *os.File, bufferCapacity int) (*sql.DB, error) {
	db, err := sql.Open("postgres", "")
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(dumpFile)
	scanner.Buffer(make([]byte, bufferCapacity), bufferCapacity)

	db.Exec("DROP TABLE IF EXISTS lsif_annotate;")
	db.Exec("CREATE TABLE lsif_annotate (data jsonb);")

	for scanner.Scan() {
		_, err := db.Exec("INSERT INTO lsif_annotate VALUES ($1);", scanner.Text())
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}
