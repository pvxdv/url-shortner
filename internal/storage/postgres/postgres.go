package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"url-shortener/internal/config"
	"url-shortener/internal/storage"
)

const (
	createTableUrlSQL = `
    CREATE TABLE IF NOT EXISTS url(
        id SERIAL PRIMARY KEY,
        alias TEXT NOT NULL UNIQUE,
        url TEXT NOT NULL);
    CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
    `
)

type Storage struct {
	db *pgx.Conn
}

func New(ctx context.Context, cfg *config.Config) (*Storage, error) {
	const fn = "storage.postgres.New"

	cs := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)

	conn, err := pgx.Connect(context.TODO(), cs)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect database:%w", fn, err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%s: failed to ping  database:%w", fn, err)
	}

	if _, err = conn.Exec(ctx, createTableUrlSQL); err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("%s: failed create url table:%w", fn, err)
	}

	return &Storage{db: conn}, nil
}

func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error) {
	const fn = "storage.postgres.SaveUrl"

	stmt, err := s.db.Prepare(ctx, "saveUrl", "INSERT INTO url(url, alias) VALUES($1,$2) RETURNING id")
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	var id int64

	err = s.db.QueryRow(ctx, stmt.Name, urlToSave, alias).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return 0, fmt.Errorf("%s: %w", fn, storage.ErrURLExists)
			}
		}
		return 0, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) GetUrl(ctx context.Context, alias string) (string, error) {
	const fn = "storage.postgres.GetUrl"

	stmt, err := s.db.Prepare(ctx, "getUrl", "SELECT url FROM url WHERE alias = $1")
	if err != nil {
		return "", fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	var url string

	err = s.db.QueryRow(ctx, stmt.Name, alias).Scan(&url)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return "", fmt.Errorf("%s: %w", fn, storage.ErrURLNotFound)
			}
			return "", fmt.Errorf("%s: failed to exec statement: %w", fn, err)
		}
	}

	return url, nil
}

func (s *Storage) DeleteUrl(ctx context.Context, alias string) error {
	const fn = "storage.postgres.DeleteUrl"

	stmt, err := s.db.Prepare(ctx, "deleteUrl", "DELETE FROM url WHERE alias = $1")
	if err != nil {
		return fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	_, err = s.db.Exec(ctx, stmt.Name, alias)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("%s: %w", fn, storage.ErrURLNotFound)
			}
			return fmt.Errorf("%s: failed to exec statement: %w", fn, err)
		}
	}

	return nil
}

func (s *Storage) UpdateUrl(ctx context.Context, urlToUpdate string, alias string) (int64, error) {
	const fn = "storage.postgres.GetUrl"

	stmt, err := s.db.Prepare(ctx, "getUrl", "UPDATE url SET url=$1 WHERE alias = $2 RETURNING id")
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", fn, err)
	}

	var id int64

	err = s.db.QueryRow(ctx, stmt.Name, urlToUpdate, alias).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, fmt.Errorf("%s: %w", fn, storage.ErrURLNotFound)
			}
			return 0, fmt.Errorf("%s: failed to exec statement: %w", fn, err)
		}
	}

	return id, nil
}
