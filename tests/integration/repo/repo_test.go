package repo_test

import (
	"backend-trainee-assignment-2024/internal/config"
	"backend-trainee-assignment-2024/internal/errs"
	"backend-trainee-assignment-2024/internal/models"
	n "backend-trainee-assignment-2024/internal/nullable"
	"backend-trainee-assignment-2024/internal/repo"
	"backend-trainee-assignment-2024/internal/storage/postgres"
	"context"
	"encoding/json"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BannerRepositoryTestSuite struct {
	suite.Suite
	db   *postgres.Postgres
	repo *repo.BannerRepository
}

func TestBannerRepository(t *testing.T) {
	suite.Run(t, new(BannerRepositoryTestSuite))
}

func (s *BannerRepositoryTestSuite) SetupSuite() {
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("failed init config: %v", err)
	}

	s.db, err = postgres.New(cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

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

	s.repo = repo.NewBannerRepository(s.db)
}

func (s *BannerRepositoryTestSuite) TearDownSuite() {
	Cleanup(s.db)
	s.db.Close()
}

func (s *BannerRepositoryTestSuite) Test0_GetBanner() {
	s.Run("GetBanner", func() {
		tests := []struct {
			name        string
			tagID       int64
			featureID   int64
			onlyActive  bool
			expected    models.Banner
			expectedErr error
		}{
			{
				name:       "successful get banner",
				tagID:      2001,
				featureID:  1001,
				onlyActive: true,
				expected: models.Banner{
					ID:        1,
					Content:   json.RawMessage(`{"title": "some_title_2", "text": "some_text", "url": "some_url"}`),
					IsActive:  true,
					CreatedAt: 1672531200,
					UpdatedAt: 1672531320,
				},
				expectedErr: nil,
			},
			{
				name:        "banner not found",
				tagID:       1,
				featureID:   1,
				onlyActive:  true,
				expected:    models.Banner{},
				expectedErr: errs.ErrNotFound,
			},
			{
				name:        "inactive banner",
				tagID:       2002,
				featureID:   1002,
				onlyActive:  true,
				expected:    models.Banner{},
				expectedErr: errs.ErrNotFound,
			},
		}

		for _, test := range tests {
			s.Run(test.name, func() {
				banner, err := s.repo.GetBanner(context.Background(), test.tagID, test.featureID, test.onlyActive)

				if test.expectedErr != nil {
					assert.ErrorIs(s.T(), err, test.expectedErr)
					assert.Equal(s.T(), test.expected, banner)
					return
				}
				require.NoError(s.T(), err)
				assert.Equal(s.T(), test.expected.ID, banner.ID)
				assert.JSONEq(s.T(), string(test.expected.Content), string(banner.Content))
				assert.Equal(s.T(), test.expected.IsActive, banner.IsActive)
				assert.Equal(s.T(), test.expected.CreatedAt, banner.CreatedAt)
				assert.Equal(s.T(), test.expected.UpdatedAt, banner.UpdatedAt)
			})
		}
	})
}

func (s *BannerRepositoryTestSuite) Test1_GetBannerByID() {
	s.Run("GetBannerByID", func() {
		tests := []struct {
			name        string
			id          int64
			expected    models.Banner
			expectedErr error
		}{
			{
				name: "successful get banner by ID",
				id:   1,
				expected: models.Banner{
					ID:        1,
					Content:   json.RawMessage(`{"title": "some_title_2", "text": "some_text", "url": "some_url"}`),
					IsActive:  true,
					CreatedAt: 1672531200,
					UpdatedAt: 1672531320,
					FeatureID: 1007,
					TagIds:    []int64{2007, 2008},
				},
				expectedErr: nil,
			},
			{
				name:        "banner not found",
				id:          9999123,
				expected:    models.Banner{},
				expectedErr: errs.ErrNotFound,
			},
		}

		for _, test := range tests {
			s.Run(test.name, func() {
				banner, err := s.repo.GetBannerByID(context.Background(), test.id)

				if test.expectedErr != nil {
					assert.ErrorIs(s.T(), err, test.expectedErr)
					assert.Equal(s.T(), test.expected, banner)
					return
				}

				require.NoError(s.T(), err)
				assert.Equal(s.T(), test.expected.ID, banner.ID)
				assert.JSONEq(s.T(), string(test.expected.Content), string(banner.Content))
				assert.Equal(s.T(), test.expected.IsActive, banner.IsActive)
				assert.Equal(s.T(), test.expected.CreatedAt, banner.CreatedAt)
				assert.Equal(s.T(), test.expected.UpdatedAt, banner.UpdatedAt)
			})
		}
	})
}

func (s *BannerRepositoryTestSuite) Test2_ListBanners() {
	s.Run("ListBanners", func() {
		tests := []struct {
			name        string
			featureID   n.NullInt64
			tagID       n.NullInt64
			limit       n.NullUint64
			offset      n.NullUint64
			expected    []models.Banner
			expectedErr error
		}{
			{
				name:      "successful list banners",
				featureID: n.NullInt64From(1001),
				tagID:     n.NullInt64From(2001),
				limit:     n.NullUint64From(10),
				offset:    n.NullUint64From(0),
				expected: []models.Banner{
					{
						ID:        1,
						Content:   json.RawMessage(`{"title": "some_title_2", "text": "some_text", "url": "some_url"}`),
						IsActive:  true,
						CreatedAt: 1672531200,
						UpdatedAt: 1672531320,
						FeatureID: 1001,
						TagIds:    []int64{2001},
					},
				},
				expectedErr: nil,
			},
			{
				name:      "filter by feature ID",
				featureID: n.NullInt64From(1003),
				tagID:     n.NullInt64{},
				limit:     n.NullUint64{},
				offset:    n.NullUint64{},
				expected: []models.Banner{
					{
						ID:        3,
						Content:   []byte(`{"title": "title", "text": "some_text", "url": "some_url"}`),
						IsActive:  false,
						CreatedAt: 1672617600,
						UpdatedAt: 1672617600,
						FeatureID: 1003,
						TagIds:    []int64{2003},
					},
					{
						ID:        4,
						Content:   []byte(`{"title": "title", "text": "some_text", "url": "some_url"}`),
						IsActive:  false,
						CreatedAt: 1672617600,
						UpdatedAt: 1672617600,
						FeatureID: 1003,
						TagIds:    []int64{2004},
					},
				},
				expectedErr: nil,
			},
			{
				name:      "filter by tag ID",
				featureID: n.NullInt64{},
				tagID:     n.NullInt64From(2005),
				limit:     n.NullUint64{},
				offset:    n.NullUint64{},
				expected: []models.Banner{
					{
						ID:        5,
						Content:   []byte(`{"title": "title", "text": "some_text", "url": "some_url"}`),
						IsActive:  false,
						CreatedAt: 1672617600,
						UpdatedAt: 1672617600,
						FeatureID: 1003,
						TagIds:    []int64{2005},
					},
					{
						ID:        6,
						Content:   []byte(`{"title": "title", "text": "some_text", "url": "some_url"}`),
						IsActive:  false,
						CreatedAt: 1672617600,
						UpdatedAt: 1672617600,
						FeatureID: 1004,
						TagIds:    []int64{2005},
					},
				},
				expectedErr: nil,
			},
			{
				name:      "limit and offset",
				featureID: n.NullInt64{},
				tagID:     n.NullInt64{},
				limit:     n.NullUint64From(2),
				offset:    n.NullUint64From(1),
				expected: []models.Banner{
					{
						ID:        2,
						Content:   []byte(`{"title": "title", "text": "some_text", "url": "some_url"}`),
						IsActive:  false,
						CreatedAt: 1672617600,
						UpdatedAt: 1672617600,
						FeatureID: 1002,
						TagIds:    []int64{2002},
					},
					{
						ID:        3,
						Content:   []byte(`{"title": "title", "text": "some_text", "url": "some_url"}`),
						IsActive:  false,
						CreatedAt: 1672617600,
						UpdatedAt: 1672617600,
						FeatureID: 1003,
						TagIds:    []int64{2003},
					},
				},
				expectedErr: nil,
			},
			{
				name:      "with multiple tags",
				featureID: n.NullInt64{},
				tagID:     n.NullInt64From(2007),
				limit:     n.NullUint64{},
				offset:    n.NullUint64{},
				expected: []models.Banner{
					{
						ID:        7,
						Content:   []byte(`{"title": "title", "text": "some_text", "url": "some_url"}`),
						IsActive:  false,
						CreatedAt: 1672617600,
						UpdatedAt: 1672617600,
						FeatureID: 1007,
						TagIds:    []int64{2007, 2008},
					},
				},
				expectedErr: nil,
			},
			{
				name:        "no results found",
				featureID:   n.NullInt64From(9999),
				tagID:       n.NullInt64From(9999),
				limit:       n.NullUint64From(10),
				offset:      n.NullUint64From(0),
				expected:    nil,
				expectedErr: nil,
			},
		}

		for _, test := range tests {
			s.Run(test.name, func() {
				banners, err := s.repo.ListBanners(
					context.Background(),
					test.featureID,
					test.tagID,
					test.limit,
					test.offset,
				)

				if test.expectedErr != nil {
					assert.ErrorIs(s.T(), err, test.expectedErr)
					assert.Equal(s.T(), test.expected, banners)
					return
				}

				require.NoError(s.T(), err)
				if len(test.expected) != len(banners) {
					s.T().Fatalf("expected %d banners, got %d", len(test.expected), len(banners))
				}

				for i := range banners {
					assert.Equal(s.T(), test.expected[i].ID, banners[i].ID)
					assert.JSONEq(s.T(), string(test.expected[i].Content), string(banners[i].Content))
					assert.Equal(s.T(), test.expected[i].IsActive, banners[i].IsActive)
					assert.Equal(s.T(), test.expected[i].CreatedAt, banners[i].CreatedAt)
					assert.Equal(s.T(), test.expected[i].UpdatedAt, banners[i].UpdatedAt)
				}
			})
		}
	})
}

func (s *BannerRepositoryTestSuite) Test3_CreateBanner() {
	s.Run("CreateBanner", func() {
		tests := []struct {
			name        string
			tagIDs      []int64
			featureID   int64
			content     json.RawMessage
			isActive    bool
			expected    int64
			expectedErr error
		}{
			{
				name:        "successful create banner",
				tagIDs:      []int64{100002, 100003},
				featureID:   100001,
				content:     json.RawMessage(`{"title": "new banner", "text": "some text", "url": "some url"}`),
				isActive:    true,
				expectedErr: nil,
			},
			{
				name:        "unique constraint violation",
				tagIDs:      []int64{2001},
				featureID:   1001,
				content:     json.RawMessage(`{"title": "new banner", "text": "some text", "url": "some url"}`),
				isActive:    true,
				expectedErr: errs.ErrUniqueViolation,
			},
		}

		for _, test := range tests {
			s.Run(test.name, func() {
				id, err := s.repo.CreateBanner(
					context.Background(),
					test.tagIDs,
					test.featureID,
					test.content,
					test.isActive,
				)

				if test.expectedErr != nil {
					assert.ErrorIs(s.T(), err, test.expectedErr)
					return
				}

				require.NoError(s.T(), err)

				banner, err := s.repo.GetBannerByID(
					context.Background(),
					id,
				)
				require.NoError(s.T(), err)
				assert.JSONEq(s.T(), string(test.content), string(banner.Content))
				assert.Equal(s.T(), test.isActive, banner.IsActive)
				assert.ElementsMatch(s.T(), test.tagIDs, banner.TagIds)
				assert.Equal(s.T(), test.featureID, banner.FeatureID)

			})
		}
	})
}

func (s *BannerRepositoryTestSuite) Test4_UpdateBanner() {
	s.Run("UpdateBanner", func() {
		tests := []struct {
			name        string
			id          int64
			tagIDs      []int64
			featureID   n.NullInt64
			content     json.RawMessage
			isActive    n.NullBool
			expectedErr error
		}{
			{
				name:        "successful update banner",
				id:          9,
				tagIDs:      []int64{2009999, 2000995},
				featureID:   n.NullInt64{Int64: 1002, Valid: true},
				content:     json.RawMessage(`{"title": "updated banner", "text": "updated text", "url": "updated url"}`),
				isActive:    n.NullBool{Bool: false, Valid: true},
				expectedErr: nil,
			},
			{
				name:        "update only content",
				id:          8,
				tagIDs:      nil,
				featureID:   n.NullInt64{},
				content:     json.RawMessage(`{"title": "updated title", "text": "updated text", "url": "updated url"}`),
				isActive:    n.NullBool{},
				expectedErr: nil,
			},
			{
				name:        "update only isActive",
				id:          2,
				tagIDs:      nil,
				featureID:   n.NullInt64{},
				content:     nil,
				isActive:    n.NullBool{Bool: true, Valid: true},
				expectedErr: nil,
			},
			{
				name:        "update only featureID",
				id:          3,
				tagIDs:      nil,
				featureID:   n.NullInt64{Int64: 2343242, Valid: true},
				content:     nil,
				isActive:    n.NullBool{},
				expectedErr: nil,
			},
			{
				name:        "update only tagIDs",
				id:          3,
				tagIDs:      []int64{43444444, 455544444},
				featureID:   n.NullInt64{},
				content:     nil,
				isActive:    n.NullBool{},
				expectedErr: nil,
			},
			{
				name:        "banner not found",
				id:          9999966,
				tagIDs:      []int64{2003, 2004},
				featureID:   n.NullInt64{Int64: 1002, Valid: true},
				content:     json.RawMessage(`{"title": "updated banner", "text": "updated text", "url": "updated url"}`),
				isActive:    n.NullBool{Bool: false, Valid: true},
				expectedErr: errs.ErrNotFound,
			},
		}

		for _, test := range tests {
			s.Run(test.name, func() {
				// Получение первой версии баннера перед обновлением
				beforeUpdate := getLastVersion(s.repo, test.id)

				err := s.repo.UpdateBanner(context.Background(), test.id, test.tagIDs, test.featureID, test.content, test.isActive)
				if test.expectedErr != nil {
					assert.ErrorIs(s.T(), err, test.expectedErr)
					return
				}
				require.NoError(s.T(), err)

				// Получение первой версии баннера после обновления
				afterUpdate := getLastVersion(s.repo, test.id)

				// Проверка, что updatedAt изменился
				if test.content != nil && string(test.content) != "null" {
					assert.NotEqual(s.T(), beforeUpdate.UpdatedAt, afterUpdate.UpdatedAt)
				}

				// Получение актуального баннера
				banner, err := s.repo.GetBannerByID(
					context.Background(),
					test.id,
				)
				require.NoError(s.T(), err)

				assert.Equal(s.T(), test.id, banner.ID)
				if test.content != nil {
					assert.JSONEq(s.T(), string(test.content), string(banner.Content))
				}
				if test.isActive.Valid {
					assert.Equal(s.T(), test.isActive.Bool, banner.IsActive)
				}
				if test.featureID.Valid {
					assert.Equal(s.T(), test.featureID.Int64, banner.FeatureID)
				}

				if len(test.tagIDs) != 0 {
					if len(test.tagIDs) != len(banner.TagIds) {
						s.T().Fatalf("expected %d tags, got %d", len(test.tagIDs), len(banner.TagIds))
					}
					assert.ElementsMatch(s.T(), test.tagIDs, banner.TagIds)
				}
			})
		}
	})
}

func getLastVersion(repo *repo.BannerRepository, bannerID int64) models.BannerVersion {
	versions, err := repo.ListBannerVersions(context.Background(), bannerID)
	if err != nil || len(versions) == 0 {
		return models.BannerVersion{}
	}
	return versions[0]
}

func (s *BannerRepositoryTestSuite) Test5_DeleteBanner() {
	s.Run("DeleteBanner", func() {
		tests := []struct {
			name        string
			id          int64
			expectedErr error
		}{
			{
				name:        "successful delete banner",
				id:          8,
				expectedErr: nil,
			},
			{
				name:        "banner not found",
				id:          9999,
				expectedErr: errs.ErrNotFound,
			},
		}
		for _, test := range tests {
			s.Run(test.name, func() {
				err := s.repo.DeleteBanner(context.Background(), test.id)
				if test.expectedErr != nil {
					require.Error(s.T(), err)
					assert.ErrorIs(s.T(), err, test.expectedErr)
					return
				}
				require.NoError(s.T(), err)
			})
		}
	})
}

func (s *BannerRepositoryTestSuite) Test6_ListBannerVersions() {
	s.Run("ListBannerVersions", func() {
		tests := []struct {
			name        string
			bannerID    int64
			expected    []models.BannerVersion
			expectedErr error
		}{
			{
				name:     "successful list banner versions",
				bannerID: 1,
				expected: []models.BannerVersion{
					{
						BannerID:  1,
						Content:   json.RawMessage(`{"title": "some_title_1", "text": "some_text", "url": "some_url"}`),
						UpdatedAt: 1672531320,
					},
					{
						BannerID:  1,
						Content:   json.RawMessage(`{"title": "some_title", "text": "some_text", "url": "some_url"}`),
						UpdatedAt: 1672531310,
					},
				},
				expectedErr: nil,
			},
			{
				name:        "no versions",
				bannerID:    2,
				expected:    []models.BannerVersion{},
				expectedErr: nil,
			},
		}

		for _, test := range tests {
			s.Run(test.name, func() {
				versions, err := s.repo.ListBannerVersions(context.Background(), test.bannerID)
				if test.expectedErr != nil {
					require.Error(s.T(), err)
					assert.ErrorIs(s.T(), err, test.expectedErr)
					return
				}

				require.NoError(s.T(), err)
				if len(test.expected) != len(versions) {
					s.T().Fatalf("expected %d versions, got %d", len(test.expected), len(versions))
				}
				for i := range versions {
					assert.Equal(s.T(), test.expected[i].BannerID, versions[i].BannerID)
					assert.JSONEq(s.T(), string(test.expected[i].Content), string(versions[i].Content))
					assert.Equal(s.T(), test.expected[i].UpdatedAt, versions[i].UpdatedAt)
				}
			})
		}
	})
}

func (s *BannerRepositoryTestSuite) Test7_RestoreVersion() {
	s.Run("RestoreVersion", func() {
		tests := []struct {
			name        string
			bannerID    int64
			updatedAt   models.UnixTime
			expectedErr error
		}{
			{
				name:        "successful restore version",
				bannerID:    1,
				updatedAt:   1672531310,
				expectedErr: nil,
			},
			{
				name:        "banner version not found",
				bannerID:    2,
				updatedAt:   1616860800,
				expectedErr: errs.ErrNotFound,
			},
		}

		for _, test := range tests {
			s.Run(test.name, func() {
				version := getVersion(s.repo, test.bannerID, test.updatedAt)

				err := s.repo.RestoreVersion(context.Background(), test.bannerID, test.updatedAt)
				if test.expectedErr != nil {
					require.Error(s.T(), err)
					assert.ErrorIs(s.T(), err, test.expectedErr)
					return
				}
				if err != nil {
					s.T().Fatalf("got error: %v", err)
				}

				banner, err := s.repo.GetBannerByID(context.Background(), test.bannerID)
				assert.NoError(s.T(), err)
				if err != nil {
					s.T().Fatalf("got error: %v", err)
				}

				assert.Equal(s.T(), test.updatedAt, version.UpdatedAt)
				assert.Equal(s.T(), test.bannerID, version.BannerID)
				assert.JSONEq(s.T(), string(banner.Content), string(version.Content))
			})
		}
	})
}

func getVersion(repo *repo.BannerRepository, bannerID int64, updatedAt models.UnixTime) models.BannerVersion {
	versions, err := repo.ListBannerVersions(context.Background(), bannerID)
	if err != nil {
		return models.BannerVersion{}
	}
	for _, version := range versions {
		if version.UpdatedAt == updatedAt {
			return version
		}
	}
	return models.BannerVersion{}
}
