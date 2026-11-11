package shortener

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("link not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateLink(ctx context.Context, code, url string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx,
		`INSERT INTO links (code, original_url) VALUES ($1, $2) RETURNING id`,
		code, url,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert link: %w", err)
	}
	return id, nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (int64, string, error) {
	var id int64
	var url string
	err := r.pool.QueryRow(ctx,
		`SELECT id, original_url FROM links WHERE code = $1`,
		code,
	).Scan(&id, &url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", ErrNotFound
		}
		return 0, "", err
	}
	return id, url, nil
}

func (r *Repository) InsertClick(ctx context.Context, linkID int64, ua string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO clicks (link_id, user_agent) VALUES ($1, $2)`,
		linkID, ua,
	)
	if err != nil {
		return fmt.Errorf("insert click: %w", err)
	}
	return nil
}
