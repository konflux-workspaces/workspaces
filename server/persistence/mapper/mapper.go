package mapper

import "fmt"

var (
	ErrLabelOwnerNotFound       error = fmt.Errorf("owner label not found")
	ErrLabelDisplayNameNotFound error = fmt.Errorf("display-name label not found")
)

type Mapper struct{}

var Default = &Mapper{}
