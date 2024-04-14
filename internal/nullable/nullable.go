package nullable

import (
	"bytes"
	"encoding/json"
)

type NullString struct {
	Valid  bool
	String string
}

type NullInt64 struct {
	Valid bool
	Int64 int64
}

type NullUint64 struct {
	Valid  bool
	Uint64 uint64
}

type NullBool struct {
	Valid bool
	Bool  bool
}

func (n NullString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(n.String)
}

func (n *NullString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*n = NullString{}
		return nil
	}

	var res string

	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}

	*n = NullString{String: res, Valid: true}

	return nil
}

func (n NullInt64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(n.Int64)
}

func (n *NullInt64) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*n = NullInt64{}
		return nil
	}

	var res int64

	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}

	*n = NullInt64{Int64: res, Valid: true}

	return nil
}

func (n NullUint64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(n.Uint64)
}

func (n *NullUint64) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*n = NullUint64{}
		return nil
	}

	var res uint64

	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}

	*n = NullUint64{Uint64: res, Valid: true}

	return nil
}

func (n NullBool) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(n.Bool)
}

func (n *NullBool) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*n = NullBool{}
		return nil
	}

	var res bool

	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}

	*n = NullBool{Bool: res, Valid: true}

	return nil
}

func NullInt64From(value int64) NullInt64 {
	return NullInt64{
		Int64: value,
		Valid: true,
	}
}

func NullUint64From(value uint64) NullUint64 {
	return NullUint64{
		Uint64: value,
		Valid:  true,
	}
}
