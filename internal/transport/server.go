package transport

import (
	"backend-trainee-assignment-2024/internal/config"
	"backend-trainee-assignment-2024/internal/logger"
	"backend-trainee-assignment-2024/internal/models"
	n "backend-trainee-assignment-2024/internal/nullable"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type AuthService interface {
	ValidateToken(token string) (userType string, err error)
}

type BannerService interface {
	GetBanner(tagID, featureID int64, useLastRevision bool, onlyActive bool) (models.Banner, error)
	ListBanners(featureID, tagID n.NullInt64, limit, offset n.NullUint64) ([]models.Banner, error)
	CreateBanner(tagIDs []int64, featureID int64, content json.RawMessage, isActive bool) (int64, error)
	UpdateBanner(bannerID int64, tagIDs []int64, featureID n.NullInt64, content json.RawMessage, isActive n.NullBool) error
	DeleteBanner(bannerID int64) error
	ListBannerVersions(bannerID int64) ([]models.BannerVersion, error)
	RestoreVersion(bannerID int64, updatedAt models.UnixTime) error
}

type HTTPServer struct {
	Address       string
	Port          int
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
	logger        logger.Logger
	bannerService BannerService
	authService   AuthService
	server        *http.Server
}

func New(cfg config.HTTPServer, logger logger.Logger, bannerService BannerService, authService AuthService) *HTTPServer {
	return &HTTPServer{
		Address:       cfg.Address,
		Port:          cfg.Port,
		ReadTimeout:   cfg.ReadTimeout,
		WriteTimeout:  cfg.WriteTimeout,
		IdleTimeout:   cfg.IdleTimeout,
		logger:        logger,
		bannerService: bannerService,
		authService:   authService,
	}
}

func (s *HTTPServer) Start() chan error {
	r := s.getRouter()

	fullAddr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	s.logger.Info("starting server", slog.String("address", fullAddr))

	srv := &http.Server{
		Addr:         fullAddr,
		Handler:      r,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
		IdleTimeout:  s.IdleTimeout,
	}
	s.server = srv

	errCh := make(chan error, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			s.logger.Error("failed to start server")
			errCh <- err
		}
	}()

	s.logger.Info("server started")

	return errCh
}

func (s *HTTPServer) getRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	getBannerForUser := NewGetBannerForUser(s.bannerService, s.authService, s.logger)
	listBanners := NewListBanners(s.bannerService, s.authService, s.logger)
	createBanner := NewCreateBanner(s.bannerService, s.authService, s.logger)
	updateBanner := NewUpdateBanner(s.bannerService, s.authService, s.logger)
	deleteBanner := NewDeleteBanner(s.bannerService, s.authService, s.logger)
	listVersions := NewListVersions(s.bannerService, s.authService, s.logger)
	restoreVersion := NewRestoreVersion(s.bannerService, s.authService, s.logger)

	r.Get("/user_banner", getBannerForUser.Handle)
	r.Route("/banner", func(r chi.Router) {
		r.Get("/", listBanners.Handle)   // GET /banner
		r.Post("/", createBanner.Handle) // POST /banner

		r.Route("/{bannerID}", func(r chi.Router) {
			r.Patch("/", updateBanner.Handle)  // PATCH /banner/{bannerID}
			r.Delete("/", deleteBanner.Handle) // DELETE /banner/{bannerID}

			r.Route("/versions", func(r chi.Router) {
				r.Get("/", listVersions.Handle)                       // GET /banner/{bannerID}/versions
				r.Post("/{updatedAt}/restore", restoreVersion.Handle) // POST /banner/{bannerID}/versions/{versionID}/restore
			})
		})
	})

	return r
}

func (s *HTTPServer) Stop() error {
	if s.server == nil {
		return fmt.Errorf("server is not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("failed to stop server: " + err.Error())
		return err
	}

	s.logger.Info("server stopped")
	return nil
}
