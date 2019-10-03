# Language Server Indexing Format Testing Utilities

🚨 This implementation is still in very early stage and follows the latest LSIF specification closely.

## Language Server Index Format

This repository host Go binaries that test the output of an [LSIF](https://github.com/Microsoft/language-server-protocol/blob/master/indexFormat/specification.md) indexer.

## Quickstart

1. Download and build this program via `go get github.com/sourcegraph/lsif-test/cmd/lsif-validate`.
2. The binary `lsif-validate` should be installed into your `$GOPATH/bin` directory.
3. Make sure you have added `$GOPATH/bin` to your `$PATH` environment variable.
4. Run `lsif-validate ./path/to/data.lsif` to see if there are errors.

The validator is a beefier version of `lsif-util validate` and checks the following properties:

- Each JSON line conforms to the JSON schema specification of a vertex or edge
- Element IDs are unique
- All references of element occur after its definition
- A single metadata vertex exists and is the firsts element in the dump
- The project root is a valid URL
- Each document URI is a URL relative to the project root
- Each range vertex has sane bounds (non-negative line/character values and the ending position occurs strictly after the starting position)
- 1-to-n edges have a non-empty `inVs` array
- Edges refer to identifiers attached to the correct element type, as follows:

    | label                     | inV(s)                     | outV                | condition |
    | ------------------------- | -------------------------- | ------------------- | --------- |
    | `contains`                | `range`                    |                     | if outV is a `document` |
    | `item`                    | `range`                    |                     | |
    | `item`                    | `referenceResult`          |                     | if outV is a `referenceResult` |
    | `next`                    | `resultSet`                | `range`/`resultSet` | |
    | `textDocument/definition` | `definitionResult`         | `range`/`resultSet` | |
    | `textDocument/references` | `referenceResult`          | `range`/`resultSet` | |
    | `textDocument/hover`      | `hoverResult`              | `range`/`resultSet` | |
    | `moniker`                 | `moniker`                  | `range`/`resultSet` | |
    | `nextMoniker`             | `moniker`                  | `moniker`           | |
    | `packageInformation`      | `packageInformation`       | `moniker`           | |

- Each vertex is reachable from a range vertex (*ignored: metadata, project, document, and event vertices*)
- Each range belongs to a unique document
- No two ranges belonging to the same document overlap
- The inVs of each `item` edge belong to that document refered to by the edge's `document` field
