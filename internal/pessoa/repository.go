package pessoa

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/bytedance/sonic"
	"github.com/isadoramsouza/rinha-de-backend-go/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/rueidis"
)

var (
	ErrDuplicateApelido = errors.New("duplicate apelido")
	ErrNotFound         = errors.New("pessoa not found")
)

type Repository interface {
	Save(ctx context.Context, p domain.Pessoa) error
	Get(ctx context.Context, id string) (domain.Pessoa, error)
	SearchByTerm(ctx context.Context, t string) ([]domain.Pessoa, error)
	Count(ctx context.Context) (int, error)
}

type repository struct {
	db    *pgxpool.Pool
	cache rueidis.Client
}

func NewRepository(db *pgxpool.Pool, redis rueidis.Client) Repository {
	return &repository{
		db:    db,
		cache: redis,
	}
}

func (r *repository) Save(ctx context.Context, p domain.Pessoa) error {
	pessoaAlreadyExistInCache, err := r.PessoaByApelidoExistsInCache(p.Apelido)
	if pessoaAlreadyExistInCache {
		return ErrDuplicateApelido
	}

	_, err = r.db.Exec(ctx,
		"INSERT INTO public.pessoas(id, apelido, nome, nascimento, stack) VALUES ($1, $2, $3, $4, $5)",
		p.ID, p.Apelido, p.Nome, p.Nascimento, p.Stack)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "pessoas_apelido_key"`:
			return ErrDuplicateApelido
		default:
			return err
		}
	}

	go r.SavePessoaInCache(&p)

	return nil
}

func (r *repository) Get(ctx context.Context, id string) (domain.Pessoa, error) {
	pessoaFromCache, err1 := r.GetPessoaByIDInCache(id)

	if pessoaFromCache != nil {
		return *pessoaFromCache, nil
	}

	if err1 != nil {
		return domain.Pessoa{}, err1
	}
	query := "SELECT id,apelido,nome,nascimento,stack FROM pessoas WHERE id=$1;"
	row := r.db.QueryRow(ctx, query, id)
	p := domain.Pessoa{}
	err := row.Scan(&p.ID, &p.Apelido, &p.Nome, &p.Nascimento, &p.Stack)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return domain.Pessoa{}, ErrNotFound
		}
		return domain.Pessoa{}, err
	}
	return p, nil
}

func (r *repository) SearchByTerm(ctx context.Context, t string) ([]domain.Pessoa, error) {
	query := `SELECT id, apelido, nome, nascimento, stack FROM pessoas p
	WHERE p.BUSCA_TRGM ILIKE '%' || $1 || '%'
	LIMIT 50;`
	rows, err := r.db.Query(ctx, query, t)
	if err != nil {
		return nil, err
	}

	var pessoas []domain.Pessoa

	for rows.Next() {
		p := domain.Pessoa{}
		_ = rows.Scan(&p.ID, &p.Apelido, &p.Nome, &p.Nascimento, &p.Stack)
		pessoas = append(pessoas, p)
	}

	return pessoas, nil
}

func (r *repository) Count(ctx context.Context) (int, error) {
	query := "SELECT COUNT(id) FROM pessoas;"
	row := r.db.QueryRow(ctx, query)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *repository) SavePessoaInCache(pessoa *domain.Pessoa) error {
	ctx := context.Background()
	pessoaString, err := sonic.MarshalString(pessoa)
	if err != nil {
		return err
	}

	pessoaInCached := r.cache.B().Set().Key("pessoa-id-" + pessoa.ID).Value(pessoaString).Build()
	apelidoInCache := r.cache.B().Set().Key("pessoa-apelido-" + pessoa.Apelido).Value(pessoa.Apelido).Build()
	for _, resp := range r.cache.DoMulti(ctx, apelidoInCache, pessoaInCached) {
		if err := resp.Error(); err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) PessoaByApelidoExistsInCache(apelido string) (bool, error) {
	ctx := context.Background()
	exists, _ := r.cache.Do(ctx, r.cache.B().Get().Key("pessoa-apelido-"+apelido).Build()).ToString()
	if exists != "" {
		return true, nil
	}
	return false, nil
}

func (r *repository) GetPessoaByIDInCache(id string) (*domain.Pessoa, error) {
	ctx := context.Background()
	pessoaResult, err := r.cache.Do(ctx, r.cache.B().Get().Key("pessoa-id-"+id).Build()).ToString()
	if err != nil {
		return nil, err
	}
	var pessoa *domain.Pessoa
	err = json.Unmarshal([]byte(pessoaResult), &pessoa)
	if err != nil {
		return nil, err
	}
	return pessoa, nil
}
