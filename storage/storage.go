package storage

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"urlshortener/errs"
)

type MemStorage struct {
	shortToLong map[string]string
	longToShort map[string]string

	len atomic.Int32
	mu  sync.RWMutex
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLen = 6

func generateShortUrl(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		shortToLong: make(map[string]string),
		longToShort: make(map[string]string),
	}
}

func (s *MemStorage) SetUrl(ctx context.Context, url string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.longToShort) != 0 {
		if shortUrl, exists := s.longToShort[url]; exists {
			return shortUrl, nil
		}
	}

	shortUrl := generateShortUrl(codeLen)
	s.shortToLong[shortUrl] = url
	s.longToShort[url] = shortUrl
	s.len.Add(1)

	return shortUrl, nil

}

func (s *MemStorage) GetUrl(ctx context.Context, shortUrl string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	s.mu.RLock()
	url, exists := s.shortToLong[shortUrl]
	s.mu.RUnlock()
	if !exists {
		return "", errs.ErrCodeNotFound
	}
	return url, nil
}

func (s *MemStorage) DeleteUrl(ctx context.Context, shortUrl string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.shortToLong[shortUrl]; !exists {
		return errs.ErrURLNotFound
	} else {
		longUrl := s.shortToLong[shortUrl]
		delete(s.shortToLong, shortUrl)
		delete(s.longToShort, longUrl)
		s.len.Add(-1)
		return nil
	}

}

// func (s *MemStorage) Size() int { return int(s.len.Load()) }
