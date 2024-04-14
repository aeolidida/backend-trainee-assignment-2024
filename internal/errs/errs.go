package errs

import (
	"errors"
	"fmt"
)

var (
	ErrRequiredValue = errors.New("value is required")
	ErrRequiredToken = errors.New("token is required")
	ErrInvalidValue  = errors.New("invalid value")
	ErrScanValueType = errors.New("invalid scan value type")
	ErrInvalidJSON   = errors.New("invalid JSON")
)

var (
	ErrNotFound        = errors.New("not found")
	ErrUniqueViolation = errors.New("unique violation")
)

var (
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidExpirationTime   = errors.New("invalid expiration time")
	ErrTokenExpired            = errors.New("token is expired")
	ErrInvalidUserType         = errors.New("invalid user type")
	ErrInvalidToken            = errors.New("invalid token")
)

func Wrap(op, msg string, err error) error {
	return fmt.Errorf("%s: %s: %w", op, msg, err)
}
