package marshal

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
)
