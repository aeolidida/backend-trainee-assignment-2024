package e2e

import (
	"backend-trainee-assignment-2024/internal/storage/postgres"
	"context"
	"encoding/json"
	"os"
	"time"
)

func loadBannersFixture(db *postgres.Postgres, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var fixtures struct {
		Banners []struct {
			ID        int64           `json:"id"`
			Content   json.RawMessage `json:"content"`
			CreatedAt int64           `json:"created_at"`
			UpdatedAt int64           `json:"updated_at"`
		} `json:"banners"`
	}

	if err := json.Unmarshal(data, &fixtures); err != nil {
		return err
	}

	for _, banner := range fixtures.Banners {
		_, err := db.Exec(context.Background(),
			"INSERT INTO banners (id, content, created_at, updated_at) VALUES ($1, $2, $3, $4)",
			banner.ID, banner.Content, time.Unix(banner.CreatedAt, 0), time.Unix(banner.UpdatedAt, 0),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func loadBannersHistoryFixture(db *postgres.Postgres, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var fixtures struct {
		BannersHistory []struct {
			BannerID  int64           `json:"banner_id"`
			Content   json.RawMessage `json:"content"`
			UpdatedAt int64           `json:"updated_at"`
		} `json:"banners_history"`
	}

	if err := json.Unmarshal(data, &fixtures); err != nil {
		return err
	}

	for _, history := range fixtures.BannersHistory {
		_, err := db.Exec(context.Background(),
			"INSERT INTO banners_history (banner_id, content, updated_at) VALUES ($1, $2, $3)",
			history.BannerID, history.Content, time.Unix(history.UpdatedAt, 0),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func loadBannerMappingsFixture(db *postgres.Postgres, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var fixtures struct {
		BannerMappings []struct {
			BannerID  int64 `json:"banner_id"`
			FeatureID int64 `json:"feature_id"`
			TagID     int64 `json:"tag_id"`
			IsActive  bool  `json:"is_active"`
		} `json:"banner_mappings"`
	}

	if err := json.Unmarshal(data, &fixtures); err != nil {
		return err
	}

	for _, mapping := range fixtures.BannerMappings {
		_, err := db.Exec(context.Background(),
			"INSERT INTO banner_mappings (banner_id, feature_id, tag_id, is_active) VALUES ($1, $2, $3, $4)",
			mapping.BannerID, mapping.FeatureID, mapping.TagID, mapping.IsActive,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func Cleanup(db *postgres.Postgres) error {
	_, err := db.Exec(context.Background(), "DELETE FROM banners")
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), "DELETE FROM banners_history")
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), "DELETE FROM banner_mappings")
	return err
}
