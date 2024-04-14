package banner

import (
	"backend-trainee-assignment-2024/internal/config"
	"backend-trainee-assignment-2024/internal/logger"
	"backend-trainee-assignment-2024/internal/models"
	n "backend-trainee-assignment-2024/internal/nullable"

	"context"
	"encoding/json"
	"time"
)

type BannerRepository interface {
	GetBanner(ctx context.Context, tagID, featureID int64, onlyActive bool) (models.Banner, error)
	GetBannerByID(ctx context.Context, id int64) (models.Banner, error)
	ListBanners(ctx context.Context, featureID, tagID n.NullInt64, limit, offset n.NullUint64) ([]models.Banner, error)
	CreateBanner(ctx context.Context, tagIDs []int64, featureID int64, content json.RawMessage, isActive bool) (int64, error)
	UpdateBanner(ctx context.Context, bannerID int64, tagIDs []int64, featureID n.NullInt64, content json.RawMessage, isActive n.NullBool) error
	DeleteBanner(ctx context.Context, bannerID int64) error
	ListBannerVersions(ctx context.Context, bannerID int64) ([]models.BannerVersion, error)
	RestoreVersion(ctx context.Context, bannerID int64, updatedAt models.UnixTime) error
}

type Cache interface {
	Push(key, value string, ttl time.Duration) (err error)
	Get(key string) (value string, err error)
	Remove(key string) (err error)
}

type Queue interface {
	Publish(message []byte, queue string) error
	Consume(queue string) (<-chan []byte, error)
	Close() error
}

type BannerService struct {
	CacheTTL         time.Duration
	BannerRepository BannerRepository
	Cache            *BannerCache
	Queue            *BannerQueue
	workerStopCh     chan struct{}
}

func NewBannerService(cfg config.BannerService, bannerRepository BannerRepository, cache Cache, queue Queue, logger logger.Logger) *BannerService {
	bannerQueue := NewBannerQueue(queue, cfg.QueueName)
	worker, workerStopCh := NewDeleteBannerWorker(bannerRepository, bannerQueue, logger)

	for i := 0; i < cfg.DeleteWorkersNum; i++ {
		go runWorker(context.Background(), worker, workerStopCh, logger)
	}

	return &BannerService{
		CacheTTL:         cfg.CacheTTL,
		BannerRepository: bannerRepository,
		Cache:            NewBannerCache(cache),
		Queue:            bannerQueue,
		workerStopCh:     workerStopCh,
	}
}

func (s *BannerService) Shutdown() {
	close(s.workerStopCh)
}

func runWorker(ctx context.Context, worker *DeleteBannerWorker, stopCh chan struct{}, logger logger.Logger) {
	for {
		err := worker.Run(ctx, stopCh)
		if err != nil {
			logger.Error("Error running worker: %v", err)
		}

		select {
		case <-stopCh:
			return
		case <-time.After(time.Second * 5):
			// Ожидание 5 секунд перед перезапуском
		}
	}
}

func (s *BannerService) GetBanner(tagID, featureID int64, useLastRevision bool, onlyActive bool) (models.Banner, error) {
	if useLastRevision {
		banner, err := s.BannerRepository.GetBanner(context.Background(), tagID, featureID, onlyActive)
		if err != nil {
			return models.Banner{}, err
		}
		go s.Cache.Push(banner, s.CacheTTL)
		return banner, nil
	}

	banner, err := s.Cache.Get(tagID, featureID)
	if err == nil {
		return banner, nil
	}

	banner, err = s.BannerRepository.GetBanner(context.TODO(), tagID, featureID, onlyActive)
	if err != nil {
		return models.Banner{}, err
	}

	go s.Cache.Push(banner, s.CacheTTL)
	return banner, nil
}

func (s *BannerService) ListBanners(featureID, tagID n.NullInt64, limit, offset n.NullUint64) ([]models.Banner, error) {
	return s.BannerRepository.ListBanners(context.TODO(), featureID, tagID, limit, offset)
}

func (s *BannerService) CreateBanner(tagIDs []int64, featureID int64, content json.RawMessage, isActive bool) (int64, error) {
	return s.BannerRepository.CreateBanner(context.TODO(), tagIDs, featureID, content, isActive)
}

func (s *BannerService) UpdateBanner(bannerID int64, tagIDs []int64, featureID n.NullInt64, content json.RawMessage, isActive n.NullBool) error {
	return s.BannerRepository.UpdateBanner(context.TODO(), bannerID, tagIDs, featureID, content, isActive)
}

func (s *BannerService) ListBannerVersions(bannerID int64) ([]models.BannerVersion, error) {
	return s.BannerRepository.ListBannerVersions(context.TODO(), bannerID)
}

func (s *BannerService) RestoreVersion(bannerID int64, updatedAt models.UnixTime) error {
	return s.BannerRepository.RestoreVersion(context.TODO(), bannerID, updatedAt)
}

func (s *BannerService) DeleteBanner(bannerID int64) error {
	_, err := s.BannerRepository.GetBannerByID(context.Background(), bannerID)
	if err != nil {
		return err
	}

	task := &DeleteBannerTask{
		ID: bannerID,
	}
	err = s.Queue.Publish(task)
	if err != nil {
		return err
	}
	return nil
}
