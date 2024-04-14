package models

import (
	"backend-trainee-assignment-2024/internal/errs"
	"database/sql/driver"
	"encoding/json"
)

type Banner struct {
	ID        int64           `json:"id"`
	Content   json.RawMessage `json:"content"`
	IsActive  bool            `json:"is_active"`
	CreatedAt UnixTime        `json:"created_at"`
	UpdatedAt UnixTime        `json:"updated_at"`
	FeatureID int64           `json:"feature_id"`
	TagIds    []int64         `json:"tag_ids"`
}

func (b *Banner) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errs.ErrScanValueType
	}
	return json.Unmarshal(bytes, &b.Content)
}

func (b Banner) Value() (driver.Value, error) {
	if b.Content == nil {
		return nil, nil
	}
	return json.Marshal(b.Content)
}
