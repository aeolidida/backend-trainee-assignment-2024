package postgres

import (
	"backend-trainee-assignment-2024/internal/config"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type Postgres struct {
	*pgxpool.Pool
}

func New(cfg config.Postgres) (*Postgres, error) {
	conn_string := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)
	fmt.Println(conn_string)

	pg := &Postgres{}

	poolConfig, err := pgxpool.ParseConfig(conn_string)
	if err != nil {
		return nil, fmt.Errorf("database.NewPostgres error in pgxpool.ParseConfig: %w", err)
	}

	var db *pgxpool.Pool
	db, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err == nil {
		err = db.Ping(context.Background())
	}
	if err == nil {
		pg.Pool = db
		return pg, pg.Ping()
	}

	return nil, errors.New("database.NewPostgres unable to connect to Postgres database")
}

func (p *Postgres) Ping() error {
	if p.Pool == nil {
		return nil
	}
	err := p.Pool.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("database.Postgres.Ping error: %w", err)
	}
	return nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

func IfErrNoRows(err error) bool {
	return err == pgx.ErrNoRows
}

func IfUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
