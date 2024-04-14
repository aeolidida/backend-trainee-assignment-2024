package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"backend-trainee-assignment-2024/internal/errs"
	"backend-trainee-assignment-2024/internal/logger"
	n "backend-trainee-assignment-2024/internal/nullable"

	"github.com/go-chi/chi/v5"
)

type GetBannerForUser struct {
	BannerService BannerService
	AuthService   AuthService
	logger        logger.Logger
}

func ErrorResponse(msg string) string {
	type ErrorResponseBody struct {
		Error string `json:"error"`
	}
	body, _ := json.Marshal(ErrorResponseBody{Error: msg})
	return string(body)
}

func NewGetBannerForUser(bannerService BannerService, authService AuthService, logger logger.Logger) *GetBannerForUser {
	return &GetBannerForUser{
		BannerService: bannerService,
		AuthService:   authService,
		logger:        logger,
	}
}

func (h *GetBannerForUser) validate(tagIDStr, featureIDStr, useLastRevisionStr string) (int64, int64, bool, error) {
	tagID, err := validateInt64(tagIDStr, true, n.NullInt64{})
	if err != nil {
		return 0, 0, false, fmt.Errorf("validate tag id: %w", err)
	}

	featureID, err := validateInt64(featureIDStr, true, n.NullInt64{})
	if err != nil {
		return 0, 0, false, fmt.Errorf("validate feature id: %w", err)
	}

	useLastRevision, err := validateBool(useLastRevisionStr, false, n.NullBool{Valid: true, Bool: false})
	if err != nil {
		return 0, 0, false, fmt.Errorf("validate use_last_revision: %w", err)
	}

	return tagID.Int64, featureID.Int64, useLastRevision.Bool, nil
}

func (h *GetBannerForUser) Handle(w http.ResponseWriter, r *http.Request) {
	tagIDStr := r.URL.Query().Get("tag_id")
	featureIDStr := r.URL.Query().Get("feature_id")
	useLastRevisionStr := r.URL.Query().Get("use_last_revision")
	token := r.Header.Get("token")

	tagID, featureID, useLastRevision, err := h.validate(
		tagIDStr,
		featureIDStr,
		useLastRevisionStr,
	)
	if err != nil {
		http.Error(w, ErrorResponse(err.Error()), http.StatusBadRequest)
		return
	}

	userType, err := h.AuthService.ValidateToken(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if useLastRevision && userType != AdminRole {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var onlyActive bool

	if userType == UserRole {
		onlyActive = true
	}

	banner, err := h.BannerService.GetBanner(tagID, featureID, useLastRevision, onlyActive)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			h.logger.Error(err.Error())
			http.Error(w, ErrorResponse(InternalServerErrorMsg), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(banner.Content)
}

type ListBanners struct {
	BannerService BannerService
	AuthService   AuthService
	logger        logger.Logger
}

func NewListBanners(bannerService BannerService, authService AuthService, logger logger.Logger) *ListBanners {
	return &ListBanners{
		BannerService: bannerService,
		AuthService:   authService,
		logger:        logger,
	}
}

func (h *ListBanners) validate(featureIDStr, tagIDStr, limitStr, offsetStr string) (n.NullInt64, n.NullInt64, n.NullUint64, n.NullUint64, error) {
	tagID, err := validateInt64(tagIDStr, false, n.NullInt64{})
	if err != nil {
		return n.NullInt64{}, n.NullInt64{}, n.NullUint64{}, n.NullUint64{}, fmt.Errorf("validate tag_id: %w", err)
	}

	featureID, err := validateInt64(featureIDStr, false, n.NullInt64{})
	if err != nil {
		return n.NullInt64{}, n.NullInt64{}, n.NullUint64{}, n.NullUint64{}, fmt.Errorf("validate feature_id: %w", err)
	}

	limit, err := validateUint64(limitStr, false, n.NullUint64{})
	if err != nil {
		return n.NullInt64{}, n.NullInt64{}, n.NullUint64{}, n.NullUint64{}, fmt.Errorf("validate limit: %w", err)
	}
	if limit.Valid && limit.Uint64 == 0 {
		return n.NullInt64{}, n.NullInt64{}, n.NullUint64{}, n.NullUint64{}, fmt.Errorf("limit can't be 0: %w", errs.ErrInvalidValue)
	}

	offset, err := validateUint64(offsetStr, false, n.NullUint64{})
	if err != nil {
		return n.NullInt64{}, n.NullInt64{}, n.NullUint64{}, n.NullUint64{}, fmt.Errorf("validate offset: %w", err)
	}

	return featureID, tagID, limit, offset, nil
}

func (h *ListBanners) Handle(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	featureID, tagID, limit, offset, err := h.validate(
		r.URL.Query().Get("feature_id"),
		r.URL.Query().Get("tag_id"),
		r.URL.Query().Get("limit"),
		r.URL.Query().Get("offset"),
	)
	if err != nil {
		http.Error(w, ErrorResponse(err.Error()), http.StatusBadRequest)
		return
	}

	userType, err := h.AuthService.ValidateToken(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if userType != AdminRole {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	banners, err := h.BannerService.ListBanners(featureID, tagID, limit, offset)
	if err != nil {
		h.logger.Error(err.Error())
		http.Error(w, InternalServerErrorMsg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	if len(banners) > 0 {
		json.NewEncoder(w).Encode(banners)
	}
}

type CreateBanner struct {
	BannerService BannerService
	AuthService   AuthService
	logger        logger.Logger
}

func NewCreateBanner(bannerService BannerService, authService AuthService, logger logger.Logger) *CreateBanner {
	return &CreateBanner{
		BannerService: bannerService,
		AuthService:   authService,
		logger:        logger,
	}
}

func (h *CreateBanner) Handle(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	userType, err := h.AuthService.ValidateToken(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if userType != AdminRole {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	type bannerData struct {
		TagIDs    []int64         `json:"tag_ids"`
		FeatureID int64           `json:"feature_id"`
		Content   json.RawMessage `json:"content"`
		IsActive  bool            `json:"is_active"`
	}

	var banner bannerData
	if err := json.NewDecoder(r.Body).Decode(&banner); err != nil {
		http.Error(w, ErrorResponse(err.Error()), http.StatusBadRequest)
		return
	}

	if banner.Content == nil || string(banner.Content) == "null" {
		http.Error(w, ErrorResponse("content cannot be nil"), http.StatusBadRequest)
		return
	}
	if len(banner.TagIDs) == 0 {
		http.Error(w, ErrorResponse("tag_ids array cannot be empty"), http.StatusBadRequest)
		return
	}

	bannerID, err := h.BannerService.CreateBanner(banner.TagIDs, banner.FeatureID, banner.Content, banner.IsActive)
	if err != nil {
		h.logger.Error(err.Error())
		http.Error(w, ErrorResponse(InternalServerErrorMsg), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"banner_id": bannerID})
}

type UpdateBanner struct {
	BannerService BannerService
	AuthService   AuthService
	logger        logger.Logger
}

func NewUpdateBanner(bannerService BannerService, authService AuthService, logger logger.Logger) *UpdateBanner {
	return &UpdateBanner{
		BannerService: bannerService,
		AuthService:   authService,
		logger:        logger,
	}
}

func (h *UpdateBanner) validate(idStr string) (int64, error) {
	id, err := validateInt64(idStr, true, n.NullInt64{})
	if err != nil {
		return 0, fmt.Errorf("validate bannerID: %w", err)
	}

	return id.Int64, nil
}

func (h *UpdateBanner) Handle(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	type bannerPatchData struct {
		TagIDs    []int64         `json:"tag_ids"`
		FeatureID n.NullInt64     `json:"feature_id"`
		Content   json.RawMessage `json:"content"`
		IsActive  n.NullBool      `json:"is_active"`
	}

	bannerID, err := h.validate(
		chi.URLParam(r, "bannerID"),
	)

	if err != nil {
		http.Error(w, ErrorResponse(err.Error()), http.StatusBadRequest)
		return
	}

	userType, err := h.AuthService.ValidateToken(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if userType != AdminRole {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var bannerData bannerPatchData

	if err := json.NewDecoder(r.Body).Decode(&bannerData); err != nil {
		http.Error(w, ErrorResponse(err.Error()), http.StatusBadRequest)
		return
	}

	if err := h.BannerService.UpdateBanner(bannerID, bannerData.TagIDs, bannerData.FeatureID, bannerData.Content, bannerData.IsActive); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else if errors.Is(err, errs.ErrUniqueViolation) {
			http.Error(w, ErrorResponse(ConflictMsg), http.StatusBadRequest)
		} else {
			h.logger.Error(err.Error())
			http.Error(w, ErrorResponse(InternalServerErrorMsg), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

type DeleteBanner struct {
	BannerService BannerService
	AuthService   AuthService
	logger        logger.Logger
}

func NewDeleteBanner(bannerService BannerService, authService AuthService, logger logger.Logger) *DeleteBanner {
	return &DeleteBanner{
		BannerService: bannerService,
		AuthService:   authService,
		logger:        logger,
	}
}

func (h *DeleteBanner) validate(idStr string) (int64, error) {
	id, err := validateInt64(idStr, true, n.NullInt64{})
	if err != nil {
		return 0, fmt.Errorf("validate bannerID: %w", err)
	}

	return id.Int64, nil
}

func (h *DeleteBanner) Handle(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	bannerID, err := h.validate(
		chi.URLParam(r, "bannerID"),
	)
	if err != nil {
		http.Error(w, ErrorResponse(err.Error()), http.StatusBadRequest)
		return
	}

	userType, err := h.AuthService.ValidateToken(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if userType != AdminRole {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := h.BannerService.DeleteBanner(bannerID); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			h.logger.Error(err.Error())
			http.Error(w, ErrorResponse(InternalServerErrorMsg), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
