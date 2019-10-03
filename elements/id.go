package elements

import (
	"encoding/json"
	"fmt"
)

type ID struct {
	Value string
}

func (id ID) String() string {
	return id.Value
}

func (fi *ID) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		return fi.unmarshalStringID(b)
	}

	return fi.unmarshalIntegerID(b)
}

func (fi *ID) unmarshalStringID(b []byte) error {
	var v string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	fi.Value = fmt.Sprintf(`"%s"`, v)
	return nil
}

func (fi *ID) unmarshalIntegerID(b []byte) error {
	var v int
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	fi.Value = fmt.Sprintf("%d", v)
	return nil
}
