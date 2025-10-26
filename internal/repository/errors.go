package repository

import "errors"

var (
	ErrUnauthenticated   = errors.New("unauthenticated")
	ErrForbidden         = errors.New("forbidden")
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("conflict")
	ErrPromoCodeNotFound = errors.New("promo code not found")
)
