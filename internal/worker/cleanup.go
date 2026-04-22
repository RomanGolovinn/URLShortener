package worker

import (
	"context"
	"log"
	"time"
)

type CleanerRepo interface {
	DeleteExpired(ctx context.Context, cutoff time.Time) (int64, error)
}

func StartCleaner(ctx context.Context, repo CleanerRepo, checkInterval time.Duration, expiration time.Duration) {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	log.Println("Фоновый воркер очистки БД запущен")

	for {
		select {
		case <-ctx.Done():
			log.Println("Фоновый воркер очистки БД остановлен")
			return

		case <-ticker.C:
			cutoff := time.Now().Add(-expiration)

			deleted, err := repo.DeleteExpired(context.Background(), cutoff)
			if err != nil {
				log.Printf("Ошибка при очистке БД: %v", err)
				continue
			}

			if deleted > 0 {
				log.Printf("Очистка завершена. Удалено старых ссылок: %d", deleted)
			}
		}
	}
}
