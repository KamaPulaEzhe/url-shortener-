package storage

import "context"

type Storage interface {
	SetUrl(ctx context.Context, url string) (string, error)
	GetUrl(ctx context.Context, shortUrl string) (string, error)
	DeleteUrl(ctx context.Context, shortUrl string) error
}
