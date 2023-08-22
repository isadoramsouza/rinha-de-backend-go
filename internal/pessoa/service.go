package pessoa

import (
	"context"

	"github.com/isadoramsouza/rinha-de-backend-go/internal/domain"
)

type Service interface {
	Save(ctx context.Context, p domain.Pessoa) (domain.Pessoa, error)
	SearchByTerm(ctx context.Context, t string) ([]domain.Pessoa, error)
	Get(ctx context.Context, id string) (domain.Pessoa, error)
	Count(ctx context.Context) (int, error)
}

type pessoaService struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &pessoaService{
		repository: r,
	}
}

func (s *pessoaService) Save(ctx context.Context, p domain.Pessoa) (domain.Pessoa, error) {
	err := s.repository.Save(ctx, p)

	return p, err
}

func (s *pessoaService) SearchByTerm(ctx context.Context, t string) ([]domain.Pessoa, error) {
	pessoas, err := s.repository.SearchByTerm(ctx, t)
	return pessoas, err
}

func (s *pessoaService) Get(ctx context.Context, id string) (domain.Pessoa, error) {
	pessoa, err := s.repository.Get(ctx, id)
	return pessoa, err
}

func (s *pessoaService) Count(ctx context.Context) (int, error) {
	count, err := s.repository.Count(ctx)

	if err != nil {
		return 0, err
	}
	return count, nil
}
