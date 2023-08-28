package formatting

import (
	"encoding/json"
)

type JSONDocument map[string]interface{}

func SerializeJSON(doc JSONDocument) ([]byte, error) {
	return json.Marshal(doc)
}

func DeserializeJSON(data []byte) (JSONDocument, error) {
	var doc JSONDocument
	err := json.Unmarshal(data, &doc)
	return doc, err
}