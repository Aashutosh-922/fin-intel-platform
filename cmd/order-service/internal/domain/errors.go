package domain

import "errors"

var (
	ErrInvalidQuantity = errors.New("invalid quantity")
	ErrInvalidPrice    = errors.New("invalid price")
	ErrDuplicateOrder  = errors.New("duplicate order")
)