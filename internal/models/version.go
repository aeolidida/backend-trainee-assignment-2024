package models

import "encoding/json"

type BannerVersion struct {
	BannerID  int64           `json:"banner_id"`
	Content   json.RawMessage `json:"content"`
	UpdatedAt UnixTime        `json:"updated_at"`
}
