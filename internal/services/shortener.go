package services

import "context"

// Shortener defines business operations for creating and resolving short URLs.
type Shortener interface {
	Create(ctx context.Context, original string) (code string, err error)
	Resolve(ctx context.Context, code string) (original string, err error)
}

// ErrNotFound is returned when a code cannot be resolved.
var ErrNotFound = &NotFoundError{}

type NotFoundError struct{}

func (NotFoundError) Error() string { return "not found" }
