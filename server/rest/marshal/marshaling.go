package marshal

import "net/http"

type MarshalerProvider func(r *http.Request) (Marshaler, error)
type UnmarshalerProvider func(r *http.Request) (Unmarshaler, error)

type Marshaler interface {
	ContentType() string
	Marshal(any) ([]byte, error)
}

type Unmarshaler interface {
	ContentType() string
	Unmarshal([]byte, any) error
}

var (
	DefaultMarshal   Marshaler   = &JsonMarshaler{}
	DefaultUnmarshal Unmarshaler = &JsonUnmarshaler{}

	DefaultMarshalerProvider   MarshalerProvider   = func(r *http.Request) (Marshaler, error) { return DefaultMarshal, nil }
	DefaultUnmarshalerProvider UnmarshalerProvider = func(r *http.Request) (Unmarshaler, error) { return DefaultUnmarshal, nil }
)
