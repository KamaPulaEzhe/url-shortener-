package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"urlshortener/storage"
)

func TestGetUrlHandler(t *testing.T) {
	tests := map[string]struct {
		Url    string
		status int
	}{
		"okey":      {"https://google.com", http.StatusPermanentRedirect},
		"not found": {"", http.StatusNotFound},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			server := NewUrlServer(storage.NewMemStorage())
			mux := newMux(server)

			code := "doesnotexist"
			if tc.Url != "" {
				seeded, err := server.store.SetUrl(context.Background(), tc.Url)
				if err != nil {
					t.Fatalf("seed failed: %v", err)
				}
				code = seeded
			}

			r := httptest.NewRequest("GET", "/"+code, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)

			if w.Code != tc.status {
				t.Fatalf("got %d, want %d", w.Code, tc.status)
			}
			if tc.status == http.StatusPermanentRedirect {
				if got := w.Header().Get("Location"); got != tc.Url {
					t.Fatalf("got Location %q, want %q", got, tc.Url)
				}
			}
		})
	}
}

func TestSetUrlHandler(t *testing.T) {
	tests := map[string]struct {
		body   string
		status int
	}{
		"okey":     {`{"url": "https://google.com"}`, http.StatusOK},
		"err url":  {`{"url": "SisiPisi"}`, http.StatusBadRequest},
		"err json": {"sosok", http.StatusBadRequest},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			server := NewUrlServer(storage.NewMemStorage())
			mux := newMux(server)

			r := httptest.NewRequest("POST", "/shorten/", strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)

			if w.Code != tc.status {
				t.Fatalf("got %d, want %d", w.Code, tc.status)
			}
			if tc.status == http.StatusOK {
				if len(w.Body.String()) != 6 {
					t.Fatalf("expected 6-char code, got %q", w.Body.String())
				}
			}

		})
	}
}

func TestDeleteUrlHandler(t *testing.T) {
	tests := map[string]struct {
		Url    string
		status int
	}{
		"okey":      {"https://google.com", http.StatusNoContent},
		"not found": {"", http.StatusNotFound},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			server := NewUrlServer(storage.NewMemStorage())
			mux := newMux(server)

			code := "doesnotexist"
			if tc.Url != "" {
				seeded, err := server.store.SetUrl(context.Background(), tc.Url)
				if err != nil {
					t.Fatalf("seed failed: %v", err)
				}
				code = seeded
			}

			r := httptest.NewRequest("DELETE", "/"+code, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)

			if w.Code != tc.status {
				t.Fatalf("got %d, want %d", w.Code, tc.status)
			}
			if tc.status == http.StatusNoContent {
				r2 := httptest.NewRequest("GET", "/"+code, nil)
				w2 := httptest.NewRecorder()
				mux.ServeHTTP(w2, r2)
				if w2.Code != http.StatusNotFound {
					t.Fatalf("code still resolгоves after delete: got %d", w2.Code)
				}
			}
		})
	}
}
