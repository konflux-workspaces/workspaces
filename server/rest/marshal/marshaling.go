package marshal

import "encoding/json"

var (
	DefaultMarshal   MarshalFunc   = json.Marshal
	DefaultUnmarshal UnmarshalFunc = json.Unmarshal
)

type MarshalFunc func(any) ([]byte, error)

type UnmarshalFunc func([]byte, any) error
