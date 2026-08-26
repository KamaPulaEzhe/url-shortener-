package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"urlshortener/errs"
)

type countingStorage struct {
	Storage
	getCalls int
}

func (c *countingStorage) GetUrl(ctx context.Context, shortUrl string) (string, error) {
	c.getCalls++
	return c.Storage.GetUrl(ctx, shortUrl)
}

func TestCachedStorage_SetUrl(t *testing.T) {
	ctxEnd, cancel := context.WithTimeout(context.Background(), 0)
	cancel()

	tests := map[string]struct {
		ctx     context.Context
		url     string
		wantErr error
	}{
		"okey":       {context.Background(), "https://example.com", nil},
		"ctx cancel": {ctxEnd, "https://example.com", ctxEnd.Err()},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			c := NewCachedStorage(NewMemStorage())
			code, err := c.SetUrl(tc.ctx, tc.url)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("got err %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if len(code) != 6 {
				t.Fatalf("expected 6-char code, got %q", code)
			}
		})
	}
}

func TestCachedStorage_GetUrl(t *testing.T) {
	t.Run("hit avoids the underlying store and returns the right value", func(t *testing.T) {
		backend := &countingStorage{Storage: NewMemStorage()}
		c := NewCachedStorage(backend)

		code, err := c.SetUrl(context.Background(), "https://example.com")
		if err != nil {
			t.Fatalf("SetUrl: %v", err)
		}

		// первый GetUrl — miss, идёт в backend и кладёт результат в кэш
		if _, err := c.GetUrl(context.Background(), code); err != nil {
			t.Fatalf("first GetUrl: %v", err)
		}
		callsAfterFirst := backend.getCalls

		// повторные обращения должны браться из кэша — backend больше не трогаем
		for i := 0; i < 3; i++ {
			got, err := c.GetUrl(context.Background(), code)
			if err != nil {
				t.Fatalf("GetUrl #%d: %v", i, err)
			}
			if got != "https://example.com" {
				t.Fatalf("GetUrl #%d: got %q, want the real URL", i, got)
			}
		}

		if backend.getCalls != callsAfterFirst {
			t.Fatalf("cache hit still called the underlying storage: got %d calls, want %d", backend.getCalls, callsAfterFirst)
		}
	})

	t.Run("not found propagates", func(t *testing.T) {
		c := NewCachedStorage(NewMemStorage())
		_, err := c.GetUrl(context.Background(), "doesnotexist")
		if !errors.Is(err, errs.ErrCodeNotFound) {
			t.Fatalf("got %v, want ErrCodeNotFound", err)
		}
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctxEnd, cancel := context.WithTimeout(context.Background(), 0)
		cancel()
		c := NewCachedStorage(NewMemStorage())
		_, err := c.GetUrl(ctxEnd, "whatever")
		if !errors.Is(err, ctxEnd.Err()) {
			t.Fatalf("got %v, want %v", err, ctxEnd.Err())
		}
	})
}

func TestCachedStorage_DeleteUrl(t *testing.T) {
	t.Run("deletes a cached entry from both layers", func(t *testing.T) {
		c := NewCachedStorage(NewMemStorage())
		code, _ := c.SetUrl(context.Background(), "https://example.com")
		if _, err := c.GetUrl(context.Background(), code); err != nil {
			t.Fatalf("seed GetUrl: %v", err)
		}

		if err := c.DeleteUrl(context.Background(), code); err != nil {
			t.Fatalf("DeleteUrl: %v", err)
		}
		if _, err := c.GetUrl(context.Background(), code); !errors.Is(err, errs.ErrCodeNotFound) {
			t.Fatalf("got %v, want ErrCodeNotFound", err)
		}
	})

	t.Run("deletes an entry that exists in the backend but was never cached", func(t *testing.T) {
		mem := NewMemStorage()
		c := NewCachedStorage(mem)
		code, _ := mem.SetUrl(context.Background(), "https://never-cached.example.com") // в обход кэша

		if err := c.DeleteUrl(context.Background(), code); err != nil {
			t.Fatalf("DeleteUrl on uncached-but-real entry: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		c := NewCachedStorage(NewMemStorage())
		err := c.DeleteUrl(context.Background(), "doesnotexist")
		if !errors.Is(err, errs.ErrURLNotFound) {
			t.Fatalf("got %v, want ErrURLNotFound", err)
		}
	})
}

func TestCachedStorage_Cleanup(t *testing.T) {
	c := NewCachedStorage(NewMemStorage())

	oldCode, _ := c.SetUrl(context.Background(), "https://old.example.com")
	freshCode, _ := c.SetUrl(context.Background(), "https://fresh.example.com")

	if v, ok := c.recent.Load(oldCode); ok {
		e := v.(cacheEntry)
		e.expiresAt = time.Now().Add(-10 * time.Minute)
		c.recent.Store(oldCode, e)
	}

	c.cleanup(3 * time.Minute)

	if _, stillCached := c.recent.Load(oldCode); stillCached {
		t.Error("old entry should have been evicted")
	}
	if _, stillCached := c.recent.Load(freshCode); !stillCached {
		t.Error("fresh entry should NOT have been evicted")
	}
}
