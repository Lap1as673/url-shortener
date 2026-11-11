package stats

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("code not found")

type Stats struct {
	Code        string `json:"code"`
	OriginalURL string `json:"original_url"`
	TotalClicks int64  `json:"total_clicks"`
	Last24h     int64  `json:"last_24h"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetStats(ctx context.Context, code string) (*Stats, error) {
	var s Stats
	err := r.pool.QueryRow(ctx, `
		SELECT l.code, l.original_url,
		       COUNT(c.id) FILTER (WHERE c.id IS NOT NULL) AS total,
		       COUNT(c.id) FILTER (WHERE c.clicked_at >= now() - interval '24 hours') AS last24h
		FROM links l
		LEFT JOIN clicks c ON c.link_id = l.id
		WHERE l.code = $1
		GROUP BY l.id
	`, code).Scan(&s.Code, &s.OriginalURL, &s.TotalClicks, &s.Last24h)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}
