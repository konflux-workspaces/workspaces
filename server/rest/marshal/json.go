package marshal

import "encoding/json"

const ContentTypeJson string = "application/json"

type JsonEncoder struct {
	JsonMarshaler
	JsonUnmarshaler
}

type JsonMarshaler struct{}

func (m *JsonMarshaler) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (m *JsonMarshaler) ContentType() string {
	return ContentTypeJson
}

type JsonUnmarshaler struct{}

func (j *JsonUnmarshaler) Unmarshal(d []byte, r any) error {
	return json.Unmarshal(d, r)
}

func (m *JsonUnmarshaler) ContentType() string {
	return ContentTypeJson
}
