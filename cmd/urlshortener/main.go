package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/RomanGolovinn/urlshortener/internal/handler"
	"github.com/RomanGolovinn/urlshortener/internal/repository"
	"github.com/RomanGolovinn/urlshortener/internal/service"
	"github.com/RomanGolovinn/urlshortener/internal/worker"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются переменные системы")
	}

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL не задан")
	}

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	repo := repository.NewPostgresRepo(pool)

	lenStr := os.Getenv("URL_LEN")
	length, err := strconv.Atoi(lenStr)
	if err != nil {
		log.Fatal("Неверный формат длинны короткой ссылки")
	}
	gen := service.NewRandomGenerator(length)
	svc := service.NewService(gen, repo)

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080/"
	}

	h := handler.NewHandler(svc, baseURL)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", h.ShortenURL)
	mux.HandleFunc("GET /{code}", h.Redirect)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Printf("Сервер запущен на порту %s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	workerCtx, workerCancel := context.WithCancel(context.Background())
	go worker.StartCleaner(workerCtx, repo, 24*time.Hour, 30*24*time.Hour)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	workerCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при остановке сервера: %v", err)
	}

	pool.Close()
}
