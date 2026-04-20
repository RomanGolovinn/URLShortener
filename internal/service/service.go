package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/RomanGolovinn/urlshortener/internal/repository"
)

const maxRetries = 5

type Service struct {
	gen  Generator
	repo repository.Repo
}

func NewService(gen Generator, repo repository.Repo) *Service {
	return &Service{
		gen:  gen,
		repo: repo,
	}
}

func (s *Service) ShortenURL(ctx context.Context, url string) (string, error) {
	createdAt := time.Now()
	for i := 0; i < maxRetries; i++ {
		shortCode, err := s.gen.Generate()
		if err != nil {
			return "", fmt.Errorf("ошибка генерации кода: %w", err)
		}
		err = s.repo.Save(ctx, url, shortCode, createdAt)
		if err != nil {
			if errors.Is(err, repository.ErrCollision) {
				continue
			}
			return "", fmt.Errorf("ошибка сохранения в БД: %w", err)
		}
		return shortCode, nil
	}
	return "", fmt.Errorf("не удалось сгенерировать уникальный код после %d попыток", maxRetries)
}

func (s *Service) ProcessRedirect(ctx context.Context, shortCode string) (string, error) {
	url, err := s.getLink(ctx, shortCode)
	if err != nil {
		return "", err
	}

	bgCtx := context.WithoutCancel(ctx)

	go func() {
		if err := s.recordTransition(bgCtx, shortCode); err != nil {
			log.Printf("ошибка записи клика для %s: %v", shortCode, err)
		}
	}()

	return url, nil
}

func (s *Service) getLink(ctx context.Context, shortCode string) (string, error) {
	url, err := s.repo.Get(ctx, shortCode)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *Service) recordTransition(ctx context.Context, shortCode string) error {
	err := s.repo.UpdateTransition(ctx, shortCode, time.Now())
	if err != nil {
		return fmt.Errorf("ошибка обновления статистики переходов: %w", err)
	}
	return nil
}
