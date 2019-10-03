# JSON schema

This is generated from the protocol as defined in the [lsif-node](https://github.com/microsoft/lsif-node) repository. This contains only definitions, so the following trailer was appended to make it a functional schema:

```json
  "oneOf": [
    {"$ref": "#/definitions/Vertex"},
    {"$ref": "#/definitions/Edge"}
  ]
```

### Todo list

- [ ] generate this schema within this repo
