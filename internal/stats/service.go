package stats

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, code string) (*Stats, error) {
	return s.repo.GetStats(ctx, code)
}
