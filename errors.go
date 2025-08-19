package sqlcompose

import "fmt"

// ErrInvalidClause is returned when a clause of an unknown type is encountered.
type ErrInvalidClause struct {
	Clause string
}

func (e *ErrInvalidClause) Error() string {
	return fmt.Sprintf("sqlcompose: clause %q is invalid", e.Clause)
}

// NewErrInvalidClause constructs a new ErrInvalidClause for the given clause name.
func NewErrInvalidClause(clause string) error {
	return &ErrInvalidClause{Clause: clause}
}

// ErrMisplacedClause is returned when a clause is used in an invalid position.
type ErrMisplacedClause struct {
	Clause string
}

func (e *ErrMisplacedClause) Error() string {
	return fmt.Sprintf("sqlcompose: clause %q cannot be used in this position", e.Clause)
}

// NewErrMisplacedClause constructs a new ErrMisplacedClause for the given clause name.
func NewErrMisplacedClause(clause string) error {
	return &ErrMisplacedClause{Clause: clause}
}
