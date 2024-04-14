package transport

import (
	"backend-trainee-assignment-2024/internal/errs"
	n "backend-trainee-assignment-2024/internal/nullable"
	"reflect"
	"testing"
)

func TestValidateString(t *testing.T) {
	testCases := []struct {
		name         string
		value        string
		required     bool
		defaultValue n.NullString
		expected     n.NullString
		expectedErr  error
	}{
		{"Required, value provided, no default", "test", true, n.NullString{}, n.NullString{Valid: true, String: "test"}, nil},
		{"Required, value provided, with default", "test", true, n.NullString{Valid: true, String: "default"}, n.NullString{Valid: true, String: "test"}, nil},
		{"Required, value missing, no default", "", true, n.NullString{}, n.NullString{}, errs.ErrRequiredValue},
		{"Required, value missing, with default", "", true, n.NullString{Valid: true, String: "default"}, n.NullString{Valid: true, String: "default"}, nil},
		{"Optional, value provided, no default", "test", false, n.NullString{}, n.NullString{Valid: true, String: "test"}, nil},
		{"Optional, value provided, with default", "test", false, n.NullString{Valid: true, String: "default"}, n.NullString{Valid: true, String: "test"}, nil},
		{"Optional, value missing, no default", "", false, n.NullString{}, n.NullString{}, nil},
		{"Optional, value missing, with default", "", false, n.NullString{Valid: true, String: "default"}, n.NullString{Valid: true, String: "default"}, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validateString(tc.value, tc.required, tc.defaultValue)
			if err != tc.expectedErr {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestValidateInt64(t *testing.T) {
	testCases := []struct {
		name         string
		value        string
		required     bool
		defaultValue n.NullInt64
		expected     n.NullInt64
		expectedErr  error
	}{
		{"Required, value provided", "123", true, n.NullInt64{}, n.NullInt64{Valid: true, Int64: 123}, nil},
		{"Required, value missing", "", true, n.NullInt64{}, n.NullInt64{}, errs.ErrRequiredValue},
		{"Required, invalid value", "abc", true, n.NullInt64{}, n.NullInt64{}, errs.ErrInvalidValue},
		{"Optional, value provided", "123", false, n.NullInt64{Valid: true, Int64: 456}, n.NullInt64{Valid: true, Int64: 123}, nil},
		{"Optional, value missing", "", false, n.NullInt64{Valid: true, Int64: 456}, n.NullInt64{Valid: true, Int64: 456}, nil},
		{"Optional, invalid value", "abc", false, n.NullInt64{Valid: true, Int64: 456}, n.NullInt64{}, errs.ErrInvalidValue},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validateInt64(tc.value, tc.required, tc.defaultValue)
			if err != tc.expectedErr {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestValidateUint64(t *testing.T) {
	testCases := []struct {
		name         string
		value        string
		required     bool
		defaultValue n.NullUint64
		expected     n.NullUint64
		expectedErr  error
	}{
		{"Required, value provided", "123", true, n.NullUint64{}, n.NullUint64{Valid: true, Uint64: 123}, nil},
		{"Required, value missing", "", true, n.NullUint64{}, n.NullUint64{}, errs.ErrRequiredValue},
		{"Required, invalid value", "abc", true, n.NullUint64{}, n.NullUint64{}, errs.ErrInvalidValue},
		{"Optional, value provided", "123", false, n.NullUint64{Valid: true, Uint64: 456}, n.NullUint64{Valid: true, Uint64: 123}, nil},
		{"Optional, value missing", "", false, n.NullUint64{Valid: true, Uint64: 456}, n.NullUint64{Valid: true, Uint64: 456}, nil},
		{"Optional, invalid value", "abc", false, n.NullUint64{Valid: true, Uint64: 456}, n.NullUint64{}, errs.ErrInvalidValue},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validateUint64(tc.value, tc.required, tc.defaultValue)
			if err != tc.expectedErr {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestValidateBool(t *testing.T) {
	testCases := []struct {
		name         string
		value        string
		required     bool
		defaultValue n.NullBool
		expected     n.NullBool
		expectedErr  error
	}{
		{"Required, value provided (true)", "true", true, n.NullBool{}, n.NullBool{Valid: true, Bool: true}, nil},
		{"Required, value provided (false)", "false", true, n.NullBool{Valid: true, Bool: true}, n.NullBool{Valid: true, Bool: false}, nil},
		{"Required, value missing", "", true, n.NullBool{}, n.NullBool{}, errs.ErrRequiredValue},
		{"Required, invalid value", "abc", true, n.NullBool{}, n.NullBool{}, errs.ErrInvalidValue},
		{"Optional, value provided (true)", "true", false, n.NullBool{Valid: true, Bool: false}, n.NullBool{Valid: true, Bool: true}, nil},
		{"Optional, value provided (false)", "false", false, n.NullBool{Valid: true, Bool: true}, n.NullBool{Valid: true, Bool: false}, nil},
		{"Optional, value missing", "", false, n.NullBool{Valid: true, Bool: true}, n.NullBool{Valid: true, Bool: true}, nil},
		{"Optional, invalid value", "abc", false, n.NullBool{Valid: true, Bool: true}, n.NullBool{}, errs.ErrInvalidValue},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := validateBool(tc.value, tc.required, tc.defaultValue)
			if err != tc.expectedErr {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}
