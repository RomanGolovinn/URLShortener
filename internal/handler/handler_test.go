package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RomanGolovinn/urlshortener/internal/service"
)

type mockGenerator struct {
	code string
}

func (m *mockGenerator) Generate() (string, error) {
	return m.code, nil
}

type mockRepo struct {
	saveErr error
	getURL  string
	getErr  error
}

func (m *mockRepo) Save(ctx context.Context, url, shortCode string, createdAt time.Time) error {
	return m.saveErr
}

func (m *mockRepo) Get(ctx context.Context, shortCode string) (string, error) {
	return m.getURL, m.getErr
}

func (m *mockRepo) UpdateTransition(ctx context.Context, shortCode string, t time.Time) error {
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, shortCode string) error {
	return nil
}

func TestHandler_ShortenURL_Success(t *testing.T) {
	repo := &mockRepo{saveErr: nil}
	gen := &mockGenerator{code: "abc123"}

	svc := service.NewService(gen, repo)

	h := NewHandler(svc, "http://localhost:8080/")

	body := bytes.NewBufferString(`{"url": "https://google.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/shorten", body)
	w := httptest.NewRecorder()

	h.ShortenURL(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusCreated {
		t.Errorf("ожидался статус 201 Created, получили %d", res.StatusCode)
	}

	expectedBody := `{"short_url":"http://localhost:8080/abc123"}` + "\n"
	if w.Body.String() != expectedBody {
		t.Errorf("ожидалось тело %s, получили %s", expectedBody, w.Body.String())
	}
}

func TestHandler_Redirect_Success(t *testing.T) {
	repo := &mockRepo{getURL: "https://google.com", getErr: nil}
	gen := &mockGenerator{code: "abc123"}

	svc := service.NewService(gen, repo)
	h := NewHandler(svc, "http://localhost:8080/")

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{code}", h.Redirect)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusFound {
		t.Errorf("ожидался статус 302, получили %d", res.StatusCode)
	}

	location := res.Header.Get("Location")
	if location != "https://google.com" {
		t.Errorf("ожидался Location 'https://google.com', получили '%s'", location)
	}
}
