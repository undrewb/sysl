// Code generated by sysl DO NOT EDIT.
package simple

import (
	"time"

	"github.com/anz-bank/sysl-go/validator"
	"github.com/rickb777/date"
)

// Reference imports to suppress unused errors
var _ = time.Parse

// Reference imports to suppress unused errors
var _ = date.Parse

// Welcome ...
type Welcome struct {
	Content string `json:"Content"`
}

// GetRequest ...
type GetRequest struct {
}

// GetFoobarListRequest ...
type GetFoobarListRequest struct {
}

// *Welcome validator
func (s *Welcome) Validate() error {
	return validator.Validate(s)
}
