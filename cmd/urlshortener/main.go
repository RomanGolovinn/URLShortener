package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/RomanGolovinn/urlshortener/internal/handler"
	"github.com/RomanGolovinn/urlshortener/internal/repository"
	"github.com/RomanGolovinn/urlshortener/internal/service"
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
	defer pool.Close()

	repo := repository.NewPostgresRepo(pool)
	lenStr := os.Getenv("URL_LEN")
	len, err := strconv.Atoi(lenStr)
	if err != nil {
		log.Fatal("Неверный формат длинны короткой ссылки")
	}
	gen := service.NewRandomGenerator(len)
	svc := service.NewService(gen, repo)

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080/"
	}

	h := handler.NewHandler(svc, baseURL)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", h.ShortenURL)
	mux.HandleFunc("GET /{code}", h.Redirect)

	serverAddr := os.Getenv("APP_PORT")
	if serverAddr == "" {
		serverAddr = ":8080"
	}

	log.Printf("Сервер запущен на %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
