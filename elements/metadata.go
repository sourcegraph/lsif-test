package elements

import "encoding/json"

type MetaData struct {
	*Element
	ProjectRoot string `json:"projectRoot"`
}

func ParseMetaData(line string) (*MetaData, error) {
	metaData := &MetaData{}
	if err := json.Unmarshal([]byte(line), &metaData); err != nil {
		return nil, err
	}

	return metaData, nil
}
