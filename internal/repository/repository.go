package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("link not found")
	ErrCollision = errors.New("short code already exists")
)

type Repo interface {
	Get(ctx context.Context, shortCode string) (string, error)
	Save(ctx context.Context, url, shortCode string, createdAt time.Time) error
	Delete(ctx context.Context, shortCode string) error
	UpdateTransition(ctx context.Context, shortCode string, lastTransition time.Time) error
}

type PostgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (p *PostgresRepo) Get(ctx context.Context, shortCode string) (string, error) {
	q := `select url from links where short_code = $1`
	var url string
	err := p.db.QueryRow(ctx, q, shortCode).Scan(&url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return url, err
}

func (p *PostgresRepo) Save(ctx context.Context, url, shortCode string,
	createdAt time.Time) error {
	q := `insert into links (url, short_code, created_at,
	number_of_transitions, last_transition)
	values ($1, $2, $3, $4, $5)`
	_, err := p.db.Exec(ctx, q, url, shortCode, createdAt, 0, nil)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 - код ошибки Unique Violation в Postgres
			return ErrCollision
		}
		return err
	}
	return err
}

func (p *PostgresRepo) UpdateTransition(ctx context.Context, shortCode string,
	lastTransition time.Time) error {
	q := `update links set number_of_transitions = number_of_transitions + 1,
	last_transition = $1 where short_code = $2`
	_, err := p.db.Exec(ctx, q, lastTransition, shortCode)
	return err
}

func (p *PostgresRepo) Delete(ctx context.Context, shortCode string) error {
	q := `delete from links where short_code = $1`
	_, err := p.db.Exec(ctx, q, shortCode)
	return err
}
