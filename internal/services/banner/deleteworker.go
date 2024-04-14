package banner

import (
	"backend-trainee-assignment-2024/internal/logger"
	"context"
	"fmt"
	"time"
)

type DeleteBannerWorker struct {
	bannerRepository BannerRepository
	queue            *BannerQueue
	stopCh           chan struct{}
	logger           logger.Logger
}

func NewDeleteBannerWorker(bannerRepository BannerRepository, queue *BannerQueue, logger logger.Logger) (*DeleteBannerWorker, chan struct{}) {
	stopCh := make(chan struct{})
	return &DeleteBannerWorker{
		bannerRepository: bannerRepository,
		queue:            queue,
		stopCh:           stopCh,
		logger:           logger,
	}, stopCh
}

func (w *DeleteBannerWorker) Run(ctx context.Context, stopCh chan struct{}) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-stopCh:
			return nil
		default:
			task, err := w.queue.Consume()
			if err != nil {
				w.logger.Error("error consuming from queue: %v", err)
				continue
			}

			err = w.deleteBanner(ctx, task)
			if err != nil {
				w.logger.Error("error deleting banner: %v", err)
				if err := w.retryDeleteBanner(ctx, task); err != nil {
					w.logger.Error("error retrying delete banner: %v", err)
					continue
				}
			}
		}
	}
}

func (w *DeleteBannerWorker) deleteBanner(ctx context.Context, task *DeleteBannerTask) error {
	return w.bannerRepository.DeleteBanner(ctx, task.ID)
}

func (w *DeleteBannerWorker) retryDeleteBanner(ctx context.Context, task *DeleteBannerTask) error {
	var retries int
	maxRetries := 5
	for retries < maxRetries {
		err := w.deleteBanner(ctx, task)
		if err == nil {
			return nil
		}
		retries++
		err = w.queue.Publish(task)
		if err != nil {
			return err
		}
		time.Sleep(time.Second * 5) // Ожидает 5 секунд перед следующей попыткой
	}
	return fmt.Errorf("failed to delete banner after %d retries", maxRetries)
}
