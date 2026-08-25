package storage

import (
	"context"
	"errors"
	"os"
	"testing"
	"urlshortener/errs"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	godotenv.Load("../.env")
	os.Exit(m.Run())
}

func setupPgStorage(t *testing.T) *PgStorage {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	if _, err := pool.Exec(context.Background(), "TRUNCATE TABLE links"); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	return NewPgStorage(pool)
}

func TestPgStorage_SetUrl(t *testing.T) {
	ctxEnd, cancel := context.WithTimeout(context.Background(), 0)
	tests := map[string]struct {
		ctx     context.Context
		url     string
		wantErr error
	}{
		"*okey*":              {context.Background(), "https://youtube.com", nil},
		"*error ctx timeout*": {ctxEnd, "https://youtube.com", ctxEnd.Err()},
	}
	cancel()

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			s := setupPgStorage(t)
			shortCode, err := s.SetUrl(tc.ctx, tc.url)
			diffLen := (len(shortCode) == 6)
			diffErr := cmp.Diff(tc.wantErr, err)
			if tc.wantErr != nil {
				if diffErr != "" {
					t.Fatalf("%s", diffErr)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if !diffLen {
					t.Fatalf("%s", shortCode)
				}
			}
		})
	}

	t.Run("idempotent", func(t *testing.T) {
		s := setupPgStorage(t)
		code1, err := s.SetUrl(context.Background(), "https://youtube.com")
		if err != nil {
			t.Fatalf("first SetUrl: %v", err)
		}
		code2, err := s.SetUrl(context.Background(), "https://youtube.com")
		if err != nil {
			t.Fatalf("second SetUrl: %v", err)
		}
		if code1 != code2 {
			t.Fatalf("got different codes: %q vs %q", code1, code2)
		}
	})

}

func TestPgStorage_GetUrl(t *testing.T) {
	ctxEnd, cancel := context.WithTimeout(context.Background(), 0)

	tests := map[string]struct {
		ctx     context.Context
		short   string
		url     string
		wantErr error
	}{
		"ctxEnd":    {ctxEnd, "", "", ctxEnd.Err()},
		"okey":      {context.Background(), "", "https://youtube.com", nil},
		"not Found": {context.Background(), "", "", errs.ErrCodeNotFound},
	}

	cancel()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			s := setupPgStorage(t)
			shortCode := tc.short
			if tc.url != "" {
				seeded, err := s.SetUrl(context.Background(), tc.url)
				if err != nil {
					t.Fatalf("seed failed: %v", err)
				}
				shortCode = seeded
			}

			url, err := s.GetUrl(tc.ctx, shortCode)
			diff := cmp.Diff(url, tc.url)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("%s", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if diff != "" {
					t.Fatalf("%s", diff)
				}
			}
		})
	}
}

func TestPgStorage_DeleteUrl(t *testing.T) {
	ctxEnd, cancel := context.WithTimeout(context.Background(), 0)

	tests := map[string]struct {
		ctx     context.Context
		short   string
		url     string
		wantErr error
	}{
		"ctxEnd":   {ctxEnd, "", "https://youtube.com", ctxEnd.Err()},
		"okey":     {context.Background(), "", "https://youtube.com", nil},
		"notFound": {context.Background(), "", "", errs.ErrURLNotFound},
	}
	cancel()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			s := setupPgStorage(t)
			shortCode := tc.short
			if tc.url != "" {
				seeded, err := s.SetUrl(context.Background(), tc.url)
				if err != nil {
					t.Fatalf("seed failed: %v", err)
				}
				shortCode = seeded
			}

			err := s.DeleteUrl(tc.ctx, shortCode)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("%s", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if _, err := s.GetUrl(context.Background(), shortCode); !errors.Is(err, errs.ErrCodeNotFound) {
					t.Fatalf("code still resolves after delete: err=%v", err)
				}
			}
		})
	}
}
