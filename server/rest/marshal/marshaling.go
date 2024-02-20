package marshal

type MarshalFunc func(any) ([]byte, error)

type UnmarshalFunc func([]byte, any) error
