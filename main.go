package main

import (
	// "crypto/rand"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"urlshortener/errs"
	"urlshortener/middleware"
	"urlshortener/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

// func home(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("URL shortener home page"))
// }

type Storage interface {
	SetUrl(ctx context.Context, url string) (string, error)
	GetUrl(ctx context.Context, shortUrl string) (string, error)
	DeleteUrl(ctx context.Context, shortUrl string) error
}

type urlServer struct {
	store Storage
}

type urlBody struct {
	Url string `json:"url"`
}

func NewUrlServer(store Storage) *urlServer {
	return &urlServer{store: store}
}

func normalizeURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid url: %s", raw)
	}
	return u.String(), nil
}

func (s *urlServer) getUrlHandler(w http.ResponseWriter, r *http.Request) {

	short := strings.Replace(r.PathValue("code"), "\n", "", -1)
	fmt.Println("--", short, "--")
	if short == "" {
		errs.WriteError(w, errs.ErrCodeNotFound)
		fmt.Println(errs.ErrCodeNotFound.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	url, err := s.store.GetUrl(ctx, short)
	if err != nil {
		if errors.Is(err, errs.ErrCodeNotFound) {
			errs.WriteError(w, err)
			fmt.Println(err.Error())
			return
		}
	}

	http.Redirect(w, r, url, http.StatusPermanentRedirect)
	fmt.Println("get")
}

func (s *urlServer) setUrlHandler(w http.ResponseWriter, r *http.Request) {

	var u urlBody

	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		errs.WriteError(w, err)
		return
	}

	urlNorm, err := normalizeURL(u.Url)
	if err != nil {
		errs.WriteError(w, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	short, err := s.store.SetUrl(ctx, urlNorm)
	if err != nil {
		if errors.Is(err, errs.ErrAlreadyExists) {
			errs.WriteError(w, err)
			return
		}
	}
	w.Write([]byte(short))
	fmt.Println("set", short)
}

func (s *urlServer) deleteUrlHandler(w http.ResponseWriter, r *http.Request) {
	short := strings.Replace(r.PathValue("code"), "\n", "", -1)
	fmt.Println("--", short, "--")
	if short == "" {
		errs.WriteError(w, errs.ErrCodeNotFound)
		fmt.Println(errs.ErrCodeNotFound.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := s.store.DeleteUrl(ctx, short)
	if err != nil {
		errs.WriteError(w, err)
		fmt.Println(err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
	fmt.Println("delete")
}

func main() {
	mux := http.NewServeMux()
	// server := NewUrlServer(storage.NewMemStorage())
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://postgres:devpass@localhost:5433/urlshortener")
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer pool.Close()

	server := NewUrlServer(storage.NewPgStorage(pool))

	mux.HandleFunc("POST /shorten/", server.setUrlHandler)
	mux.HandleFunc("GET /{code}", server.getUrlHandler)
	mux.HandleFunc("DELETE /{code}", server.deleteUrlHandler)

	handler := middleware.PanicRecovery(mux)
	handler = middleware.Logging(handler)

	fmt.Println("Starting server at port 8080")
	log.Fatal(http.ListenAndServe("localhost:8080", handler))
}
