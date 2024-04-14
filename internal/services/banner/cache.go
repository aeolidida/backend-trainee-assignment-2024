package banner

import (
	"backend-trainee-assignment-2024/internal/models"
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

type BannerCache struct {
	cache Cache
}

func NewBannerCache(cache Cache) *BannerCache {
	return &BannerCache{
		cache: cache,
	}
}

func (bc *BannerCache) Push(banner models.Banner, ttl time.Duration) error {
	data, err := encodeBanner(banner)
	if err != nil {
		return err
	}

	for _, tagID := range banner.TagIds {
		key := fmt.Sprintf("%d_%d", tagID, banner.FeatureID)
		err := bc.cache.Push(key, string(data), ttl)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bc *BannerCache) Get(tagID, featureID int64) (models.Banner, error) {
	key := fmt.Sprintf("%d_%d", tagID, featureID)
	data, err := bc.cache.Get(key)
	if err != nil {
		return models.Banner{}, err
	}
	fmt.Println("Cache works")
	return decodeBanner([]byte(data))
}

func encodeBanner(banner models.Banner) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(banner); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeBanner(data []byte) (models.Banner, error) {
	var banner models.Banner
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&banner); err != nil {
		return models.Banner{}, err
	}
	return banner, nil
}
