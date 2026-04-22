package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RomanGolovinn/urlshortener/internal/repository"
)

type mockGenerator struct {
	code string
	err  error
}

func (m *mockGenerator) Generate() (string, error) {
	return m.code, m.err
}

type mockRepo struct {
	saveErrs  []error
	saveCalls int
	getURL    string
	getErr    error
}

func (m *mockRepo) Save(ctx context.Context, url, shortCode string, createdAt time.Time) error {
	var err error
	if m.saveCalls < len(m.saveErrs) {
		err = m.saveErrs[m.saveCalls]
	}
	m.saveCalls++
	return err
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

func TestService_ShortenURL_CollisionResolved(t *testing.T) {
	repo := &mockRepo{
		saveErrs: []error{repository.ErrCollision, repository.ErrCollision, nil},
	}
	gen := &mockGenerator{code: "aB3dE5"}
	svc := NewService(gen, repo)

	shortCode, err := svc.ShortenURL(context.Background(), "https://google.com")

	if err != nil {
		t.Fatalf("ожидалась ошибка nil, получено: %v", err)
	}
	if shortCode != "aB3dE5" {
		t.Errorf("ожидался код aB3dE5, получили %s", shortCode)
	}
	if repo.saveCalls != 3 {
		t.Errorf("ожидалось 3 вызова Save, получили %d", repo.saveCalls)
	}
}

func TestService_ShortenURL_MaxRetries(t *testing.T) {
	repo := &mockRepo{
		saveErrs: []error{
			repository.ErrCollision,
			repository.ErrCollision,
			repository.ErrCollision,
			repository.ErrCollision,
			repository.ErrCollision,
			repository.ErrCollision,
		},
	}
	gen := &mockGenerator{code: "aB3dE5"}
	svc := NewService(gen, repo)

	_, err := svc.ShortenURL(context.Background(), "https://itmo.ru")

	if err == nil {
		t.Fatal("ожидалась ошибка, но получено nil")
	}
	if repo.saveCalls != 5 {
		t.Errorf("ожидалось 5 вызовов Save, получили %d", repo.saveCalls)
	}
}

func TestService_ProcessRedirect_NotFound(t *testing.T) {
	repo := &mockRepo{
		getErr: repository.ErrNotFound,
	}
	svc := NewService(nil, repo)

	_, err := svc.ProcessRedirect(context.Background(), "unknown")

	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("ожидалась ошибка ErrNotFound, получили: %v", err)
	}
}
