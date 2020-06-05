package datastream

import "encoding/json"

type Marshaler interface {
	Marshal(v interface{}) ([]byte, error)
}

type JSONMarshaler struct{}

func (m *JSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
