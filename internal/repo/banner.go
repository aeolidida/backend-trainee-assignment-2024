package repo

import (
	"backend-trainee-assignment-2024/internal/errs"
	"backend-trainee-assignment-2024/internal/models"
	n "backend-trainee-assignment-2024/internal/nullable"
	"backend-trainee-assignment-2024/internal/storage/postgres"
	"context"
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type BannerRepository struct {
	db *postgres.Postgres
}

func NewBannerRepository(db *postgres.Postgres) *BannerRepository {
	return &BannerRepository{db: db}
}

func (r *BannerRepository) GetBanner(ctx context.Context, tagID, featureID int64, onlyActive bool) (models.Banner, error) {
	const op = "BannerRepository.GetBanner"
	builder := squirrel.Select("b.id", "b.content", "bm.is_active", "b.created_at", "b.updated_at").
		From("banners b").
		Join("banner_mappings bm ON b.id = bm.banner_id").
		Where(squirrel.Eq{
			"bm.tag_id":     tagID,
			"bm.feature_id": featureID,
		})

	if onlyActive {
		builder = builder.Where(squirrel.Eq{"bm.is_active": true})
	}

	builder = builder.PlaceholderFormat(squirrel.Dollar).Limit(1)

	query, args, err := builder.ToSql()
	if err != nil {
		return models.Banner{}, errs.Wrap(op, "failed to build SQL query", err)
	}

	var banner models.Banner

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&banner.ID, &banner.Content, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt,
	)
	if err != nil {
		if postgres.IfErrNoRows(err) {
			return models.Banner{}, errs.Wrap(op, "banner not found", errs.ErrNotFound)
		}
		return models.Banner{}, errs.Wrap(op, "failed to execute SQL query", err)
	}

	return banner, nil
}

func (r *BannerRepository) GetBannerByID(ctx context.Context, id int64) (models.Banner, error) {
	const op = "BannerRepository.GetBannerByID"

	builder := squirrel.Select("b.id", "b.content", "bm.is_active", "b.created_at", "b.updated_at", "json_agg(bm.tag_id) as tag_ids", "bm.feature_id").
		From("banners b").
		Join("banner_mappings bm ON b.id = bm.banner_id").
		Where(squirrel.Eq{"b.id": id}).
		GroupBy("b.id", "b.content", "bm.is_active", "b.created_at", "b.updated_at", "bm.feature_id").
		OrderBy("b.id").
		Limit(1)

	query, args, err := builder.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return models.Banner{}, errs.Wrap(op, "failed to build SQL query", err)
	}

	var banner models.Banner
	var tagIDsStr string

	err = r.db.QueryRow(ctx, query, args...).Scan(
		&banner.ID, &banner.Content, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt, &tagIDsStr, &banner.FeatureID,
	)
	if err != nil {
		if postgres.IfErrNoRows(err) {
			return models.Banner{}, errs.Wrap(op, "banner not found", errs.ErrNotFound)
		}
		return models.Banner{}, errs.Wrap(op, "failed to execute SQL query", err)
	}

	banner.TagIds = make([]int64, 0)
	err = json.Unmarshal([]byte(tagIDsStr), &banner.TagIds)
	if err != nil {
		return models.Banner{}, errs.Wrap(op, "failed to unmarshal tag IDs", err)
	}

	return banner, nil
}

func (r *BannerRepository) ListBanners(ctx context.Context, featureID, tagID n.NullInt64, limit, offset n.NullUint64) ([]models.Banner, error) {
	const op = "BannerRepository.ListBanners"

	builder := squirrel.Select("b.id", "b.content", "bm.is_active", "b.created_at", "b.updated_at", "json_agg(bm.tag_id) as tag_ids", "bm.feature_id").
		From("banners b").
		Join("banner_mappings bm ON b.id = bm.banner_id")

	if featureID.Valid {
		builder = builder.Where(squirrel.Eq{"bm.feature_id": featureID.Int64})
	}

	if tagID.Valid {
		builder = builder.Where(squirrel.Eq{"bm.tag_id": tagID.Int64})
	}

	if limit.Valid {
		builder = builder.Limit(limit.Uint64)
	}

	if offset.Valid {
		builder = builder.Offset(offset.Uint64)
	}

	builder = builder.GroupBy("b.id", "b.content", "bm.is_active", "b.created_at", "b.updated_at", "bm.feature_id").OrderBy("b.id")

	query, args, err := builder.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, errs.Wrap(op, "failed to build SQL query", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.Wrap(op, "failed to execute SQL query", err)
	}
	defer rows.Close()

	var banners []models.Banner
	for rows.Next() {
		var banner models.Banner

		var tagIDsStr string
		err = rows.Scan(&banner.ID, &banner.Content, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt, &tagIDsStr, &banner.FeatureID)
		if err != nil {
			return nil, errs.Wrap(op, "failed to scan row", err)
		}

		banner.TagIds = make([]int64, 0)
		err = json.Unmarshal([]byte(tagIDsStr), &banner.TagIds)
		if err != nil {
			return nil, errs.Wrap(op, "failed to unmarshal tag IDs", err)
		}
		banners = append(banners, banner)
	}

	if err = rows.Err(); err != nil {
		return nil, errs.Wrap(op, "failed to iterate over rows", err)
	}

	return banners, nil
}

func (r *BannerRepository) CreateBanner(ctx context.Context, tagIDs []int64, featureID int64, content json.RawMessage, isActive bool) (int64, error) {
	const op = "BannerRepository.CreateBanner"

	// Начало транзакции
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, errs.Wrap(op, "failed to start transaction", err)
	}
	defer tx.Rollback(ctx)

	// Вставка баннера в таблицу banners
	bannerID, err := r.insertBanner(ctx, tx, content)
	if err != nil {
		return 0, errs.Wrap(op, "failed to insert banner", err)
	}

	// Вставка связей в banner_mappings
	for _, tagID := range tagIDs {
		err = r.insertBannerMappings(ctx, tx, bannerID, featureID, tagID, isActive)
		if err != nil {
			if postgres.IfUniqueViolation(err) {
				return 0, errs.Wrap(op, "unique constraint violated", errs.ErrUniqueViolation)
			}
			return 0, errs.Wrap(op, "failed to insert banner mapping", err)
		}
	}

	// Коммит
	err = tx.Commit(ctx)
	if err != nil {
		return 0, errs.Wrap(op, "failed to commit transaction", err)
	}

	return bannerID, nil
}

func (r *BannerRepository) insertBanner(ctx context.Context, tx pgx.Tx, content json.RawMessage) (int64, error) {
	const op = "BannerRepository.insertBanner"

	query, args, err := squirrel.Insert("banners").
		Columns("content").
		Values(content).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return 0, errs.Wrap(op, "failed to build SQL query", err)
	}

	var bannerID int64
	err = tx.QueryRow(ctx, query, args...).Scan(&bannerID)
	if err != nil {
		return 0, errs.Wrap(op, "failed to insert banner", err)
	}

	return bannerID, nil
}

func (r *BannerRepository) insertBannerMappings(ctx context.Context, tx pgx.Tx, bannerID, featureID, tagID int64, isActive bool) error {
	const op = "BannerRepository.insertBannerMappings"

	query, args, err := squirrel.Insert("banner_mappings").
		Columns("banner_id", "feature_id", "tag_id", "is_active").
		Values(bannerID, featureID, tagID, isActive).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return errs.Wrap(op, "failed to build SQL query", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return errs.Wrap(op, "failed to execute SQL query", err)
	}

	return nil
}

func (r *BannerRepository) UpdateBanner(ctx context.Context, id int64, tagIDs []int64, featureID n.NullInt64, content json.RawMessage, isActive n.NullBool) error {
	const op = "BannerRepository.UpdateBanner"

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return errs.Wrap(op, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	if content != nil && string(content) != "null" {
		// Обновление записи в таблице banner
		err = r.updateBanner(ctx, tx, id, content)
		if err != nil {
			return err
		}
		// Создание новой записи в таблице banners_histpry
		err = r.createBannerHistory(ctx, tx, id, content)
		if err != nil {
			return err
		}
	}

	// Замена старых связей новыми в bannerMappings
	if len(tagIDs) > 0 || featureID.Valid || isActive.Valid {
		err = r.updateBannerMappings(ctx, tx, id, tagIDs, featureID, isActive)
		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errs.Wrap(op, "failed to commit transaction", err)
	}

	return nil
}

func (r *BannerRepository) updateBanner(ctx context.Context, tx pgx.Tx, id int64, content json.RawMessage) error {
	const op = "BannerRepository.updateBanner"

	builder := squirrel.Update("banners").Where(squirrel.Eq{"id": id})
	if content != nil || string(content) != "null" {
		builder = builder.Set("content", content)
	}

	builder = builder.Set("updated_at", time.Now())
	query, args, err := builder.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return errs.Wrap(op, "failed to build SQL query", err)
	}

	res, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return errs.Wrap(op, "failed to execute SQL query", err)
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return errs.Wrap(op, "banner not found", errs.ErrNotFound)
	}

	return nil
}

func (r *BannerRepository) createBannerHistory(ctx context.Context, tx pgx.Tx, bannerID int64, content json.RawMessage) error {
	const op = "BannerRepository.createBannerHistory"

	query, args, err := squirrel.Insert("banners_history").
		Columns("banner_id", "content", "updated_at").
		Values(bannerID, content, time.Now()).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return errs.Wrap(op, "failed to build SQL query", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return errs.Wrap(op, "failed to execute SQL query", err)
	}

	return nil
}

func (r *BannerRepository) updateBannerMappings(ctx context.Context, tx pgx.Tx, id int64, tagIDs []int64, featureID n.NullInt64, isActive n.NullBool) error {
	const op = "BannerRepository.updateBannerMappings"

	// Если ничего менять не нужно
	if !featureID.Valid && !isActive.Valid && len(tagIDs) == 0 {
		return nil
	}

	// Если не выставляем новые featureID, isActive, то надо их узнать
	// Если выставляем, то эти значения просто не используем
	var featureIDtemp int64
	var isActiveTemp bool

	if !featureID.Valid || !isActive.Valid {
		err := tx.QueryRow(ctx, "SELECT feature_id, is_active FROM banner_mappings WHERE banner_id=$1 LIMIT 1", id).Scan(&featureIDtemp, &isActiveTemp)
		if err != nil && err != pgx.ErrNoRows {
			return errs.Wrap(op, "failed to execute SQL query", err)
		}
	}
	if featureID.Valid {
		featureIDtemp = featureID.Int64
	}
	if isActive.Valid {
		isActiveTemp = isActive.Bool
	}

	// В случае, если надо обновлять теги
	if len(tagIDs) != 0 {
		// Удаление старых связей
		_, err := tx.Exec(ctx, "DELETE FROM banner_mappings WHERE banner_id=$1", id)
		if err != nil {
			return errs.Wrap(op, "failed to execute SQL query", err)
		}

		// Добавление новых связей
		for _, tagID := range tagIDs {
			_, err = tx.Exec(ctx, "INSERT INTO banner_mappings (banner_id, feature_id, tag_id, is_active) VALUES ($1, $2, $3, $4)", id, featureIDtemp, tagID, isActiveTemp)
			if err != nil {
				if postgres.IfUniqueViolation(err) {
					return errs.Wrap(op, "unique constraint violated", errs.ErrUniqueViolation)
				}
				return errs.Wrap(op, "failed to execute SQL query", err)
			}
		}
		return nil
	}

	// В случае, если не обновлялись теги
	_, err := tx.Exec(ctx, "UPDATE banner_mappings SET feature_id=$1, is_active=$2 WHERE banner_id=$3", featureIDtemp, isActiveTemp, id)
	if err != nil {
		return errs.Wrap(op, "failed to execute SQL query", err)
	}
	return nil
}

func (r *BannerRepository) DeleteBanner(ctx context.Context, id int64) error {
	const op = "BannerRepository.DeleteBanner"

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return errs.Wrap(op, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	// Удаление связей из banner_mappings
	_, err = tx.Exec(ctx, "DELETE FROM banner_mappings WHERE banner_id=$1", id)
	if err != nil {
		return errs.Wrap(op, "failed to execute SQL query", err)
	}

	// Удаление баннера из banners
	res, err := tx.Exec(ctx, "DELETE FROM banners WHERE id=$1", id)
	if err != nil {
		return errs.Wrap(op, "failed to execute SQL query", err)
	}

	// Проверка, было ли что-то удалено
	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return errs.Wrap(op, "banner not found", errs.ErrNotFound)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errs.Wrap(op, "failed to commit transaction", err)
	}

	return nil
}

func (r *BannerRepository) ListBannerVersions(ctx context.Context, bannerID int64) ([]models.BannerVersion, error) {
	const op = "BannerRepository.ListBannerVersions"
	var versions []models.BannerVersion

	builder := squirrel.Select("banner_id", "content", "updated_at").
		From("banners_history").
		Where(squirrel.Eq{
			"banner_id": bannerID,
		}).
		OrderBy("updated_at DESC")

	query, args, err := builder.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, errs.Wrap(op, "failed to build SQL query", err)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.Wrap(op, "failed to execute SQL query", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version models.BannerVersion
		if err := rows.Scan(&version.BannerID, &version.Content, &version.UpdatedAt); err != nil {
			return nil, errs.Wrap(op, "failed to scan row", err)
		}
		versions = append(versions, version)
	}

	if err := rows.Err(); err != nil {
		return nil, errs.Wrap(op, "failed to iterate over rows", err)
	}

	return versions, nil
}

func (r *BannerRepository) RestoreVersion(ctx context.Context, bannerID int64, updatedAt models.UnixTime) error {
	const op = "BannerRepository.RestoreVersion"

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return errs.Wrap(op, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	bannerVersion, err := r.deleteBannerVersion(ctx, tx, bannerID, updatedAt)
	if err != nil {
		return err
	}

	err = r.restoreBannerVersion(ctx, tx, bannerVersion)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return errs.Wrap(op, "failed to commit transaction", err)
	}

	return nil
}

func (r *BannerRepository) deleteBannerVersion(ctx context.Context, tx pgx.Tx, bannerID int64, updatedAt models.UnixTime) (models.BannerVersion, error) {
	const op = "BannerRepository.deleteBannerVersion"
	var deletedVersion models.BannerVersion

	builder := squirrel.Delete("banners_history").
		Where(squirrel.Eq{
			"banner_id":  bannerID,
			"updated_at": time.Unix(int64(updatedAt), 0),
		}).
		Suffix("RETURNING content")

	query, args, err := builder.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return models.BannerVersion{}, errs.Wrap(op, "failed to build SQL query", err)
	}

	row := tx.QueryRow(ctx, query, args...)
	err = row.Scan(&deletedVersion.Content)
	if err != nil {
		if postgres.IfErrNoRows(err) {
			return models.BannerVersion{}, errs.Wrap(op, "no banner versions", errs.ErrNotFound)

		}
		return models.BannerVersion{}, errs.Wrap(op, "failed to execute SQL query", err)
	}

	deletedVersion.BannerID = bannerID
	deletedVersion.UpdatedAt = updatedAt

	return deletedVersion, nil
}

func (r *BannerRepository) restoreBannerVersion(ctx context.Context, tx pgx.Tx, bannerVersion models.BannerVersion) error {
	const op = "BannerRepository.restoreBannerVersion"

	builder := squirrel.Update("banners").
		Set("content", bannerVersion.Content).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": bannerVersion.BannerID})

	query, args, err := builder.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return errs.Wrap(op, "failed to build SQL query", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return errs.Wrap(op, "failed to execute SQL query", err)
	}

	return nil
}
