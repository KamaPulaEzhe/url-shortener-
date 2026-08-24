package storage

import (
	"context"
	"errors"
	"urlshortener/errs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type urlForDb struct {
	shortCode string
	longCode  string
}

type PgStorage struct {
	db *pgxpool.Pool
}

func NewPgStorage(db *pgxpool.Pool) *PgStorage {
	return &PgStorage{
		db: db,
	}
}

func (s *PgStorage) SetUrl(ctx context.Context, longCode string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	shortCode := generateShortUrl(codeLen)
	_, err := s.db.Exec(ctx, "INSERT INTO links (short_code, long_url) VALUES($1, $2)", shortCode, longCode)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "links_long_url_key" {
			var url urlForDb
			err = s.db.QueryRow(ctx, "SELECT short_code, long_url FROM links WHERE long_url = $1", longCode).Scan(&url.shortCode, &url.longCode)
			if err != nil {
				return "", nil
			}
			return url.shortCode, nil
		} else if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "links_pkey" {
			for {
				_, err = s.db.Exec(ctx, "INSERT INTO links (short_code, long_url) VALUES($1, $2)", shortCode, longCode)
				errors.As(err, &pgErr)
				if err == nil {
					break
				} else if pgErr.ConstraintName == "links_pkey" {
					shortCode = generateShortUrl(codeLen)
				}

			}
		}
	}
	return shortCode, nil
}

func (s *PgStorage) GetUrl(ctx context.Context, shortCode string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	var url urlForDb
	err := s.db.QueryRow(ctx, "SELECT short_code, long_url FROM links WHERE short_code = $1", shortCode).Scan(&url.shortCode, &url.longCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrCodeNotFound
		} else {
			return "", err
		}
	}
	return url.longCode, nil
}

func (s *PgStorage) DeleteUrl(ctx context.Context, shortUrl string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	tag, err := s.db.Exec(ctx, "DELETE FROM links WHERE short_code = $1", shortUrl)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrURLNotFound
	}
	return nil
}

// func (s *PgStorage) Size() int { return int(s.len.Load())}
