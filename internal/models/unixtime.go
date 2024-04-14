package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type UnixTime int64

func (t UnixTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("%d", int64(t))
	return []byte(stamp), nil
}

func (t *UnixTime) UnmarshalJSON(data []byte) error {
	var stamp int64
	if err := json.Unmarshal(data, &stamp); err != nil {
		return err
	}
	*t = UnixTime(stamp)
	return nil
}

func (t UnixTime) Value() (time.Time, error) {
	return time.Unix(int64(t), 0), nil
}

func (t *UnixTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		*t = UnixTime(v.Unix())
	case nil:
		*t = UnixTime(0)
	default:
		return fmt.Errorf("invalid type for UnixTime: %T", value)
	}
	return nil
}
