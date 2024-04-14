package transport

import (
	"backend-trainee-assignment-2024/internal/errs"
	"backend-trainee-assignment-2024/internal/logger"
	"backend-trainee-assignment-2024/internal/models"
	n "backend-trainee-assignment-2024/internal/nullable"
	"errors"

	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ListVersions struct {
	BannerService BannerService
	AuthService   AuthService
	logger        logger.Logger
}

func NewListVersions(bannerService BannerService, authService AuthService, logger logger.Logger) *ListVersions {
	return &ListVersions{
		BannerService: bannerService,
		AuthService:   authService,
		logger:        logger,
	}
}

func (h *ListVersions) validate(idStr string) (int64, error) {
	id, err := validateInt64(idStr, true, n.NullInt64{})
	if err != nil {
		return 0, fmt.Errorf("validate id: %w", err)
	}

	return id.Int64, nil
}

func (h *ListVersions) Handle(w http.ResponseWriter, r *http.Request) {
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

	bannerVersions, err := h.BannerService.ListBannerVersions(bannerID)

	if err != nil {
		h.logger.Error(err.Error())
		http.Error(w, ErrorResponse(InternalServerErrorMsg), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if len(bannerVersions) > 0 {
		json.NewEncoder(w).Encode(bannerVersions)
	}
}

type RestoreVersion struct {
	BannerService BannerService
	AuthService   AuthService
	logger        logger.Logger
}

func NewRestoreVersion(bannerService BannerService, authService AuthService, logger logger.Logger) *RestoreVersion {
	return &RestoreVersion{
		BannerService: bannerService,
		AuthService:   authService,
		logger:        logger,
	}
}

func (h *RestoreVersion) validate(bannerIDStr, updatedAtStr string) (int64, models.UnixTime, error) {
	bannerID, err := validateInt64(bannerIDStr, true, n.NullInt64{})
	if err != nil {
		return 0, 0, fmt.Errorf("validate id: %w", err)
	}
	updatedAt, err := validateInt64(updatedAtStr, true, n.NullInt64{})
	if err != nil {
		return 0, 0, fmt.Errorf("validate id: %w", err)
	}

	return bannerID.Int64, models.UnixTime(updatedAt.Int64), nil
}

func (h *RestoreVersion) Handle(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")

	bannerID, updatedAt, err := h.validate(
		chi.URLParam(r, "bannerID"),
		chi.URLParam(r, "updatedAt"),
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

	err = h.BannerService.RestoreVersion(bannerID, updatedAt)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h.logger.Error(err.Error())
		http.Error(w, ErrorResponse(InternalServerErrorMsg), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
