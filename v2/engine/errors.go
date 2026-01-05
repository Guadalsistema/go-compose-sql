package engine

import "errors"

var (
	ErrNotInTransaction     = errors.New("connection is not in a transaction")
	ErrAlreadyInTransaction = errors.New("connection is already in a transaction")
)
