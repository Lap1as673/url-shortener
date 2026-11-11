package shortener

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

type Service struct {
	repo   *Repository
	prefix string
}

func NewService(repo *Repository, prefix string) *Service {
	return &Service{repo: repo, prefix: prefix}
}

// Shorten создаёт короткую ссылку и возвращает её полный URL.
func (s *Service) Shorten(ctx context.Context, originalURL string) (string, error) {
	originalURL = strings.TrimSpace(originalURL)
	if originalURL == "" {
		return "", fmt.Errorf("empty url")
	}

	// до 5 попыток на случай коллизии кода
	for i := 0; i < 5; i++ {
		code, err := generateCode(7)
		if err != nil {
			return "", err
		}
		_, err = s.repo.CreateLink(ctx, code, originalURL)
		if err == nil {
			return s.prefix + "/" + code, nil
		}
		// грубая проверка на дубликат уникального индекса
		if strings.Contains(err.Error(), "links_code_key") {
			continue
		}
		return "", err
	}
	return "", fmt.Errorf("could not generate unique code")
}

// Redirect ищет исходный URL и фиксирует клик.
func (s *Service) Redirect(ctx context.Context, code, userAgent string) (string, error) {
	id, url, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return "", err
	}
	// клик логируем, но не валим запрос, если не записалось
	_ = s.repo.InsertClick(ctx, id, userAgent)
	return url, nil
}

func generateCode(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// URL-safe base64, убираем паддинг и символы, которые портят URL
	code := base64.RawURLEncoding.EncodeToString(b)
	code = strings.NewReplacer("-", "a", "_", "b").Replace(code)
	if len(code) > n {
		code = code[:n]
	}
	return code, nil
}
