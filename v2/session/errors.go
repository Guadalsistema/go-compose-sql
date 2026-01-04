package session

import "errors"

var (
	// ErrNotInTransaction is returned when attempting transaction operations outside a transaction
	ErrNotInTransaction = errors.New("session is not in a transaction")

	// ErrAlreadyInTransaction is returned when attempting to start a transaction while already in one
	ErrAlreadyInTransaction = errors.New("session is already in a transaction")
)
