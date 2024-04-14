package transport

import (
	"backend-trainee-assignment-2024/internal/errs"
	n "backend-trainee-assignment-2024/internal/nullable"
	"strconv"
)

func validateString(value string, required bool, defaultValue n.NullString) (n.NullString, error) {
	if value == "" {
		if required && !defaultValue.Valid {
			return n.NullString{}, errs.ErrRequiredValue
		}
		return defaultValue, nil
	}
	return n.NullString{Valid: true, String: value}, nil
}

func validateInt64(value string, required bool, defaultValue n.NullInt64) (n.NullInt64, error) {
	if value == "" {
		if required {
			return n.NullInt64{}, errs.ErrRequiredValue
		}
		return defaultValue, nil
	}
	res, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return n.NullInt64{}, errs.ErrInvalidValue
	}
	return n.NullInt64{Valid: true, Int64: res}, nil
}

func validateUint64(value string, required bool, defaultValue n.NullUint64) (n.NullUint64, error) {
	if value == "" {
		if required {
			return n.NullUint64{}, errs.ErrRequiredValue
		}
		return defaultValue, nil
	}
	res, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return n.NullUint64{}, errs.ErrInvalidValue
	}
	return n.NullUint64{Valid: true, Uint64: res}, nil
}

func validateBool(value string, required bool, defaultValue n.NullBool) (n.NullBool, error) {
	if value == "" {
		if required {
			return n.NullBool{}, errs.ErrRequiredValue
		}
		return defaultValue, nil
	}
	res, err := strconv.ParseBool(value)
	if err != nil {
		return n.NullBool{}, errs.ErrInvalidValue
	}
	return n.NullBool{Valid: true, Bool: res}, nil
}
