package e2e

import (
	"backend-trainee-assignment-2024/internal/config"
	"backend-trainee-assignment-2024/internal/logger"
	"backend-trainee-assignment-2024/internal/models"
	"backend-trainee-assignment-2024/internal/repo"
	"backend-trainee-assignment-2024/internal/services/auth"
	"backend-trainee-assignment-2024/internal/services/banner"
	"backend-trainee-assignment-2024/internal/storage/postgres"
	"backend-trainee-assignment-2024/internal/storage/rabbitmq"
	"backend-trainee-assignment-2024/internal/storage/redis"
	"backend-trainee-assignment-2024/internal/transport"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type E2ESuite struct {
	suite.Suite
	db            *postgres.Postgres
	cache         *redis.Redis
	queue         *rabbitmq.RabbitMQQueue
	bannerService *banner.BannerService
	authService   *auth.AuthService
	server        *transport.HTTPServer
}

func TestE2ESuiteSuite(t *testing.T) {
	suite.Run(t, new(E2ESuite))
}

func (s *E2ESuite) SetupSuite() {
	// Инициализация config
	cfg, err := config.Init()
	if err != nil {
		fmt.Printf("app.Run failed to init config: %s\n", err.Error())
		os.Exit(1)
	}

	// Инициализация логгера
	logger, err := logger.New(cfg.Logger)
	if err != nil {
		fmt.Printf("app.Run failed to init logger: %s\n", err.Error())
		os.Exit(1)
	}

	// Инициализация хранилища
	logger.Info("initializing storages...")
	postgres, err := postgres.New(cfg.Postgres)
	s.db = postgres
	if err != nil {
		logger.Error(fmt.Sprintf("app.Run failed to init database: %s\n", err.Error()))
		os.Exit(1)
	}

	// Инициализация репозиториев
	logger.Info("initializing repos...")
	repo := repo.NewBannerRepository(postgres)

	cache := redis.NewRedis(cfg.Redis)
	s.cache = cache
	queue, err := rabbitmq.NewRabbitMQQueue(cfg.RabbitMQ)
	s.queue = queue

	if err != nil {
		logger.Error(fmt.Sprintf("app.Run failed to init queue: %s\n", err.Error()))
		os.Exit(1)
	}

	// Инициализация сервисов
	bannerService := banner.NewBannerService(cfg.BannerService, repo, cache, queue, logger)
	s.bannerService = bannerService
	authService := auth.NewAuthService(cfg.AuthService.SecretKey)
	s.authService = authService

	// Запуск сервера
	server := transport.New(cfg.HTTPServer, logger, bannerService, authService)
	s.server = server

	server.Start()

	time.Sleep(3 * time.Second)

	Cleanup(s.db)
	// Загрузка фикстур
	if err := loadBannersFixture(s.db, "fixtures/banners.json"); err != nil {
		s.T().Fatalf("failed to load banners fixture: %v", err)
	}

	if err := loadBannersHistoryFixture(s.db, "fixtures/versions.json"); err != nil {
		s.T().Fatalf("failed to load banners history fixture: %v", err)
	}

	if err := loadBannerMappingsFixture(s.db, "fixtures/mappings.json"); err != nil {
		s.T().Fatalf("failed to load banner mappings fixture: %v", err)
	}

}

func (s *E2ESuite) TearDownSuite() {
	s.bannerService.Shutdown()
	s.cache.Close()
	s.queue.Close()
	Cleanup(s.db)
	s.db.Close()
	s.server.Stop()
}

func (s *E2ESuite) Test0_GetBannerForUser() {
	adminToken, _ := s.authService.GenerateToken(1, "admin")
	userToken, _ := s.authService.GenerateToken(2, "user")

	tests := []struct {
		name         string
		url          string
		token        string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "successful get banner",
			url:          "/user_banner?tag_id=2001&feature_id=1001&use_last_revision=false",
			token:        adminToken,
			expectedCode: http.StatusOK,
			expectedBody: `{"title": "some_title_2", "text": "some_text", "url": "some_url"}`,
		},
		{
			name:         "invalid tag_id",
			url:          "/user_banner?tag_id=abc&feature_id=1&use_last_revision=false",
			token:        adminToken,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"validate tag id: invalid value"}`,
		},
		{
			name:         "unauthorized",
			url:          "/user_banner?tag_id=1&feature_id=1&use_last_revision=false",
			token:        "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: "",
		},
		{
			name:         "forbidden for use_last_revision",
			url:          "/user_banner?tag_id=1&feature_id=1&use_last_revision=true",
			token:        userToken,
			expectedCode: http.StatusForbidden,
			expectedBody: "",
		},
		{
			name:         "banner not found",
			url:          "/user_banner?tag_id=1&feature_id=1&use_last_revision=false",
			token:        adminToken,
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			fullAddr := fmt.Sprintf("%s:%d", s.server.Address, s.server.Port)
			path := "http://" + fullAddr + test.url

			req, err := http.NewRequest("GET", path, nil)
			assert.NoError(s.T(), err)

			req.Header.Set("token", test.token)

			cli := http.Client{}

			res, err := cli.Do(req)
			assert.NoError(s.T(), err)
			respBody, err := io.ReadAll(res.Body)
			bodyStr := string(respBody)
			defer res.Body.Close()

			assert.NoError(s.T(), err)

			assert.Equal(s.T(), test.expectedCode, res.StatusCode)

			assert.NoError(s.T(), err)

			if len(test.expectedBody) == 0 || len(bodyStr) == 0 {
				if test.expectedBody == bodyStr {
					return
				}
				s.T().Fatalf("unexpected body")
			}

			assert.JSONEq(s.T(), test.expectedBody, bodyStr)

			res.Body.Close()
		})
	}
}

func (s *E2ESuite) Test1_ListBanners() {
	adminToken, _ := s.authService.GenerateToken(1, "admin")
	userToken, _ := s.authService.GenerateToken(2, "user")

	tests := []struct {
		name         string
		url          string
		token        string
		expectedCode int
	}{
		{
			name:         "successful list banners",
			url:          "/banner?feature_id=1002&tag_id=2002&limit=10&offset=0",
			token:        adminToken,
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid feature_id",
			url:          "/banner?feature_id=abc&tag_id=2001&limit=10&offset=0",
			token:        adminToken,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid tag_id",
			url:          "/banner?feature_id=1001&tag_id=abc&limit=10&offset=0",
			token:        adminToken,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid limit",
			url:          "/banner?feature_id=1001&tag_id=2001&limit=abc&offset=0",
			token:        adminToken,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "limit can't be 0",
			url:          "/banner?feature_id=1001&tag_id=2001&limit=0&offset=0",
			token:        adminToken,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid offset",
			url:          "/banner?feature_id=1001&tag_id=2001&limit=10&offset=abc",
			token:        adminToken,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "unauthorized",
			url:          "/banner?feature_id=1001&tag_id=2001&limit=10&offset=0",
			token:        "",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "forbidden",
			url:          "/banner?feature_id=1001&tag_id=2001&limit=10&offset=0",
			token:        userToken,
			expectedCode: http.StatusForbidden,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			fullAddr := fmt.Sprintf("%s:%d", s.server.Address, s.server.Port)
			path := "http://" + fullAddr + test.url

			req, err := http.NewRequest("GET", path, nil)
			assert.NoError(s.T(), err)

			req.Header.Set("token", test.token)

			cli := http.Client{}

			res, err := cli.Do(req)
			assert.NoError(s.T(), err)

			respBody, err := io.ReadAll(res.Body)
			defer res.Body.Close()

			assert.NoError(s.T(), err)

			assert.Equal(s.T(), test.expectedCode, res.StatusCode)

			if test.expectedCode == http.StatusOK {
				var banners []models.Banner
				err = json.Unmarshal(respBody, &banners)
				assert.NoError(s.T(), err)
				assert.NotNil(s.T(), banners)
				assert.NotEmpty(s.T(), banners)
			}

			res.Body.Close()
		})
	}
}

func (s *E2ESuite) Test2_CreateBanner() {
	adminToken, _ := s.authService.GenerateToken(1, "admin")
	userToken, _ := s.authService.GenerateToken(2, "user")

	tests := []struct {
		name         string
		url          string
		token        string
		payload      string
		expectedCode int
	}{
		{
			name:         "successful create banner",
			url:          "/banner",
			token:        adminToken,
			payload:      `{"tag_ids": [2023401], "feature_id": 234234, "content": {"title": "some_title_2", "text": "some_text", "url": "some_url"}, "is_active": true}`,
			expectedCode: http.StatusCreated,
		},
		{
			name:         "invalid tag_id",
			url:          "/banner",
			token:        adminToken,
			payload:      `{"tag_ids": ["abc"], "feature_id": 1001, "content": {"title": "some_title_2", "text": "some_text", "url": "some_url"}, "is_active": true}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "content is nil",
			url:          "/banner",
			token:        adminToken,
			payload:      `{"tag_ids": [2001], "feature_id": 1001, "content": null, "is_active": true}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "tag_ids array is empty",
			url:          "/banner",
			token:        adminToken,
			payload:      `{"tag_ids": [], "feature_id": 1001, "content": {"title": "some_title_2", "text": "some_text", "url": "some_url"}, "is_active": true}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "unauthorized",
			url:          "/banner",
			token:        "",
			payload:      `{"tag_ids": [2001], "feature_id": 1001, "content": {"title": "some_title_2", "text": "some_text", "url": "some_url"}, "is_active": true}`,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "forbidden",
			url:          "/banner",
			token:        userToken,
			payload:      `{"tag_ids": [2001], "feature_id": 1001, "content": {"title": "some_title_2", "text": "some_text", "url": "some_url"}, "is_active": true}`,
			expectedCode: http.StatusForbidden,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			fullAddr := fmt.Sprintf("%s:%d", s.server.Address, s.server.Port)
			path := "http://" + fullAddr + test.url

			req, err := http.NewRequest("POST", path, strings.NewReader(test.payload))
			assert.NoError(s.T(), err)

			req.Header.Set("token", test.token)
			req.Header.Set("Content-Type", "application/json")

			cli := http.Client{}

			res, err := cli.Do(req)
			assert.NoError(s.T(), err)

			defer res.Body.Close()

			assert.Equal(s.T(), test.expectedCode, res.StatusCode)
			assert.NoError(s.T(), err)
		})
	}
}

func (s *E2ESuite) Test3_UpdateBanner() {
	adminToken, _ := s.authService.GenerateToken(1, "admin")
	userToken, _ := s.authService.GenerateToken(2, "user")

	tests := []struct {
		name         string
		url          string
		token        string
		payload      string
		expectedCode int
	}{
		{
			name:         "successful update banner",
			url:          "/banner/1",
			token:        adminToken,
			payload:      `{"tag_ids": [345346346], "feature_id": 345555, "content": {"titlenew": "some_title_updated", "text": "some_text", "url": "some_url_updated"}, "is_active": true}`,
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid tag_id",
			url:          "/banner/1",
			token:        adminToken,
			payload:      `{"tag_ids": ["abc"], "feature_id": 1001, "content": {"title": "some_title_updated", "text": "some_text", "url": "some_url_updated"}, "is_active": true}`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "banner not found",
			url:          "/banner/100555555",
			token:        adminToken,
			payload:      `{"tag_ids": [2001], "feature_id": 1001, "content": {"title": "some_title_updated", "text": "some_text", "url": "some_url_updated"}, "is_active": true}`,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "unauthorized",
			url:          "/banner/1",
			token:        "",
			payload:      `{"tag_ids": [2001], "feature_id": 1001, "content": {"title": "some_title_updated", "text": "some_text", "url": "some_url_updated"}, "is_active": true}`,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "forbidden",
			url:          "/banner/1",
			token:        userToken,
			payload:      `{"tag_ids": [2001], "feature_id": 1001, "content": {"title": "some_title_updated", "text": "some_text", "url": "some_url_updated"}, "is_active": true}`,
			expectedCode: http.StatusForbidden,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			fullAddr := fmt.Sprintf("%s:%d", s.server.Address, s.server.Port)
			path := "http://" + fullAddr + test.url

			req, err := http.NewRequest("PATCH", path, strings.NewReader(test.payload))
			assert.NoError(s.T(), err)

			req.Header.Set("token", test.token)
			req.Header.Set("Content-Type", "application/json")

			cli := http.Client{}

			res, err := cli.Do(req)
			assert.NoError(s.T(), err)
			defer res.Body.Close()

			assert.Equal(s.T(), test.expectedCode, res.StatusCode)
			assert.NoError(s.T(), err)
		})
	}
}

func (s *E2ESuite) Test4_DeleteBanner() {
	adminToken, _ := s.authService.GenerateToken(1, "admin")
	userToken, _ := s.authService.GenerateToken(2, "user")

	tests := []struct {
		name         string
		url          string
		token        string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "successful delete banner",
			url:          "/banner/1",
			token:        adminToken,
			expectedCode: http.StatusNoContent,
			expectedBody: "",
		},
		{
			name:         "banner not found",
			url:          "/banner/10680",
			token:        adminToken,
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:         "unauthorized",
			url:          "/banner/13",
			token:        "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: "",
		},
		{
			name:         "forbidden",
			url:          "/banner/13",
			token:        userToken,
			expectedCode: http.StatusForbidden,
			expectedBody: "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			fullAddr := fmt.Sprintf("%s:%d", s.server.Address, s.server.Port)
			path := "http://" + fullAddr + test.url

			req, err := http.NewRequest("DELETE", path, nil)
			assert.NoError(s.T(), err)

			req.Header.Set("token", test.token)

			cli := http.Client{}

			res, err := cli.Do(req)
			assert.NoError(s.T(), err)

			defer res.Body.Close()

			assert.Equal(s.T(), test.expectedCode, res.StatusCode)

			body, err := io.ReadAll(res.Body)
			bodyStr := string(body)
			assert.NoError(s.T(), err)

			if len(test.expectedBody) == 0 || len(body) == 0 {
				if test.expectedBody == bodyStr {
					return
				}
				s.T().Fatalf("unexpected body")
			}

			assert.JSONEq(s.T(), test.expectedBody, bodyStr)

			res.Body.Close()
		})
	}
}

func (s *E2ESuite) Test5_ListVersions() {
	adminToken, _ := s.authService.GenerateToken(1, "admin")
	userToken, _ := s.authService.GenerateToken(2, "user")

	tests := []struct {
		name         string
		url          string
		token        string
		expectedCode int
	}{
		{
			name:         "successful list versions",
			url:          "/banner/1/versions",
			token:        adminToken,
			expectedCode: http.StatusOK,
		},
		{
			name:         "unauthorized list versions",
			url:          "/banner/1/versions",
			token:        userToken,
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "invalid banner ID",
			url:          "/banner/abc/versions",
			token:        adminToken,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			fullAddr := fmt.Sprintf("%s:%d", s.server.Address, s.server.Port)
			path := "http://" + fullAddr + test.url

			req, err := http.NewRequest("GET", path, nil)
			if err != nil {
				s.T().Fatalf("failed to create request: %v", err)
			}

			req.Header.Set("token", test.token)

			cli := http.Client{}
			res, err := cli.Do(req)
			if err != nil {
				s.T().Fatalf("request failed: %v", err)
			}
			defer res.Body.Close()

			if res.StatusCode != test.expectedCode {
				s.T().Errorf("unexpected status code: got %v, want %v", res.StatusCode, test.expectedCode)
			}
		})
	}
}

func (s *E2ESuite) Test6_RestoreVersion() {
	adminToken, _ := s.authService.GenerateToken(1, "admin")
	userToken, _ := s.authService.GenerateToken(2, "user")

	tests := []struct {
		name         string
		url          string
		token        string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "successful restore version",
			url:          "/banner/1/versions/1672531320/restore",
			token:        adminToken,
			expectedCode: http.StatusOK,
		},
		{
			name:         "unauthorized restore version",
			url:          "/banner/1/versions/1672531320/restore",
			token:        userToken,
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "invalid banner ID",
			url:          "/banner/abc/versions/1672531320/restore",
			token:        adminToken,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"validate id: invalid value"}`,
		},
		{
			name:         "invalid updated at",
			url:          "/banner/1/versions/abc/restore",
			token:        adminToken,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"validate id: invalid value"}`,
		},
		{
			name:         "banner not found",
			url:          "/banner/10670/versions/1672531320/restore",
			token:        adminToken,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			fullAddr := fmt.Sprintf("%s:%d", s.server.Address, s.server.Port)
			path := "http://" + fullAddr + test.url

			req, err := http.NewRequest("POST", path, nil)
			if err != nil {
				s.T().Fatalf("failed to create request: %v", err)
			}

			req.Header.Set("token", test.token)

			cli := http.Client{}
			res, err := cli.Do(req)
			if err != nil {
				s.T().Fatalf("request failed: %v", err)
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				s.T().Fatalf("failed to read response body: %v", err)
			}
			bodyStr := string(body)

			if res.StatusCode != test.expectedCode {
				s.T().Errorf("unexpected status code: got %v, want %v", res.StatusCode, test.expectedCode)
			}

			if test.expectedBody != "" && bodyStr != "" {
				assert.JSONEq(s.T(), test.expectedBody, bodyStr)
				return
			}

			if test.expectedBody != bodyStr {
				s.T().Errorf("unexpected body: got %v, want %v", bodyStr, test.expectedBody)
			}
		})
	}
}
