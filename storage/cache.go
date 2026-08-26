package storage

import (
	"context"
	"sync"
	"time"
)

type cacheEntry struct {
	url       string
	expiresAt time.Time
}

type CachedStorage struct {
	recent  sync.Map
	storage Storage
}

func NewCachedStorage(storage Storage) *CachedStorage {
	return &CachedStorage{
		recent:  sync.Map{},
		storage: storage,
	}
}

func (c *CachedStorage) SetUrl(ctx context.Context, url string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	shortUrl, err := c.storage.SetUrl(ctx, url)
	if err != nil {
		return "", err
	}

	c.recent.Store(shortUrl, cacheEntry{url: url, expiresAt: time.Now()})

	return shortUrl, nil

}

func (c *CachedStorage) GetUrl(ctx context.Context, shortUrl string) (string, error) {
	var url string
	if err := ctx.Err(); err != nil {
		return "", err
	}

	cacheUrl, exists := c.recent.Load(shortUrl)
	if !exists {
		url, err := c.storage.GetUrl(ctx, shortUrl)
		if err != nil {
			return "", err
		}
		c.recent.Store(shortUrl, cacheEntry{url: url, expiresAt: time.Now()})
		return url, nil
	}
	c.recent.Swap(shortUrl, cacheEntry{url: cacheUrl.(cacheEntry).url, expiresAt: time.Now()})
	url = cacheUrl.(cacheEntry).url
	return url, nil
}

func (c *CachedStorage) DeleteUrl(ctx context.Context, shortUrl string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := c.storage.DeleteUrl(ctx, shortUrl); err != nil {
		return err
	}
	c.recent.LoadAndDelete(shortUrl)
	return nil
}

func (c *CachedStorage) cleanup(maxAge time.Duration) {
	c.recent.Range(func(key, value any) bool {
		if time.Since(value.(cacheEntry).expiresAt) > maxAge {
			c.recent.Delete(key)
		}
		return true
	})

}

func (c *CachedStorage) StartCleanup() {
	go func() {
		for {
			time.Sleep(time.Minute)
			c.cleanup(3 * time.Minute)
		}
	}()
}
