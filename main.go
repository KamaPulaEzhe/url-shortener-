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
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"urlshortener/errs"
	"urlshortener/middleware"
	"urlshortener/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

var rateLimit rate.Limit = 50
var rateBurst int = 100

type urlServer struct {
	store storage.Storage
}

type urlBody struct {
	Url string `json:"url"`
}

func NewUrlServer(store storage.Storage) *urlServer {
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
}

func newMux(server *urlServer) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten/", server.setUrlHandler)
	mux.HandleFunc("GET /{code}", server.getUrlHandler)
	mux.HandleFunc("DELETE /{code}", server.deleteUrlHandler)

	return mux
}

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer pool.Close()

	cacheStorage := storage.NewCachedStorage(storage.NewPgStorage(pool))
	cacheStorage.StartCleanup()

	serv := NewUrlServer(cacheStorage)
	mux := newMux(serv)

	rateLimiter := middleware.NewIPRateLimiter(rateLimit, rateBurst)
	rateLimiter.StartCleanUp()

	handler := rateLimiter.RateLimit(mux)
	handler = middleware.PanicRecovery(handler)
	handler = middleware.Logging(handler)

	server := &http.Server{Addr: "0.0.0.0:8080", Handler: handler}

	go func() {
		fmt.Println("сервер слушает :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("ошибка сервера:", err)
		}
	}()
	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Println("не успели закрыться штатно:", err)
	}
}
