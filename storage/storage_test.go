package storage

import (
	"context"
	"errors"
	"testing"
	"urlshortener/errs"

	"github.com/google/go-cmp/cmp"
)

func TestSetUrl(t *testing.T) {
	s := NewMemStorage()
	ctxEnd, cancel := context.WithTimeout(context.Background(), 0)
	tests := map[string]struct {
		ctx     context.Context
		url     string
		wantErr error
	}{
		"simple": {context.Background(), "https://youtube.com", nil},
		"error1": {ctxEnd, "https://youtube.com", ctxEnd.Err()},
	}
	cancel()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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
}

func TestGetUrl(t *testing.T) {
	s := NewMemStorage()
	ctxEnd, cancel := context.WithTimeout(context.Background(), 0)
	shortCode, _ := s.SetUrl(context.Background(), "https://youtube.com")

	tests := map[string]struct {
		ctx     context.Context
		short   string
		want    string
		wantErr error
	}{
		"ctxEnd":   {ctxEnd, "mN0hP0", "", ctxEnd.Err()},
		"okey":     {context.Background(), shortCode, "https://youtube.com", nil},
		"notFound": {context.Background(), "ghjjgh", "", errs.ErrCodeNotFound},
	}
	cancel()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			url, err := s.GetUrl(tc.ctx, tc.short)
			diff := cmp.Diff(url, tc.want)
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

func TestDeleteUrl(t *testing.T) {
	s := NewMemStorage()
	ctxEnd, cancel := context.WithTimeout(context.Background(), 0)
	shortCode, _ := s.SetUrl(context.Background(), "https://youtube.com")

	tests := map[string]struct {
		ctx     context.Context
		short   string
		wantErr error
	}{
		"ctxEnd":   {ctxEnd, "mN0hP0", ctxEnd.Err()},
		"okey":     {context.Background(), shortCode, nil},
		"notFound": {context.Background(), "ghjjgh", errs.ErrURLNotFound},
	}
	cancel()
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := s.DeleteUrl(tc.ctx, tc.short)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("%s", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
		})
	}
}
