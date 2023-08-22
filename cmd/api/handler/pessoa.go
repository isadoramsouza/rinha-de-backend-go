package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isadoramsouza/rinha-de-backend-go/internal/domain"
	"github.com/isadoramsouza/rinha-de-backend-go/internal/pessoa"
	"github.com/isadoramsouza/rinha-de-backend-go/pkg/web"
)

var (
	ErrInvalidJson = errors.New("invalid json")
	ErrNotFound    = errors.New("pessoa not found")
	InvalidDtoErr  = errors.New("invalid request")
)

type PessoaRequest struct {
	Apelido    string   `json:"apelido" validate:"required,max=32"`
	Nome       string   `json:"nome" validate:"required,max=100"`
	Nascimento string   `json:"nascimento" validate:"required,datetime=2006-01-02"`
	Stack      []string `json:"stack" validate:"dive,max=32"`
}

type PessoaResponse struct {
	ID         string   `json:"id"`
	Apelido    string   `json:"apelido" validate:"required,max=32"`
	Nome       string   `json:"nome" validate:"required,max=100"`
	Nascimento string   `json:"nascimento" validate:"required,datetime=2006-01-02"`
	Stack      []string `json:"stack" validate:"dive,max=32"`
}

func (c *PessoaRequest) Validate() error {
	if len(c.Apelido) > 32 {
		return InvalidDtoErr
	}

	if len(c.Nome) > 100 {
		return InvalidDtoErr
	}

	dateLayout := "2006-01-02"
	if _, err := time.Parse(dateLayout, c.Nascimento); err != nil {
		return InvalidDtoErr
	}

	for _, tech := range c.Stack {
		if len(tech) > 32 {
			return InvalidDtoErr
		}
	}

	return nil
}

type PessoaController struct {
	pessoaService pessoa.Service
}

func NewPessoa(s pessoa.Service) *PessoaController {
	return &PessoaController{
		pessoaService: s,
	}
}

func (p *PessoaController) SearchByTerm() gin.HandlerFunc {
	return func(c *gin.Context) {

		term := c.Query("t")

		if term == "" {
			web.Error(c, http.StatusBadRequest, "missing parameter t")
			return
		}

		pessoas, err := p.pessoaService.SearchByTerm(c, term)

		if err != nil {
			web.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		if pessoas == nil {
			web.Success(c, http.StatusOK, []PessoaResponse{})
			return
		}

		var pessoasResponse []PessoaResponse

		for _, pessoa := range pessoas {
			pessoaGet := PessoaResponse{
				ID:         pessoa.ID,
				Apelido:    pessoa.Apelido,
				Nome:       pessoa.Nome,
				Nascimento: pessoa.Nascimento,
				Stack:      strings.Split(pessoa.Stack, ","),
			}
			pessoasResponse = append(pessoasResponse, pessoaGet)
		}

		web.Success(c, http.StatusOK, pessoasResponse)
	}
}

func (p *PessoaController) Get() gin.HandlerFunc {
	return func(c *gin.Context) {
		pessoaID := fmt.Sprintf(c.Param("id"))

		pessoaGet, err := p.pessoaService.Get(c, pessoaID)
		if err != nil {
			if err.Error() == ErrNotFound.Error() {
				web.Error(c, http.StatusNotFound, ErrNotFound.Error())
				return
			}
			web.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		pessoa := PessoaResponse{
			ID:         pessoaGet.ID,
			Apelido:    pessoaGet.Apelido,
			Nome:       pessoaGet.Nome,
			Nascimento: pessoaGet.Nascimento,
			Stack:      strings.Split(pessoaGet.Stack, ","),
		}

		web.Success(c, http.StatusOK, pessoa)
	}
}

func (p *PessoaController) Create() gin.HandlerFunc {
	return func(c *gin.Context) {
		input := &PessoaRequest{}

		err := c.ShouldBindJSON(input)
		if err != nil {
			web.Error(c, http.StatusUnprocessableEntity, ErrInvalidJson.Error())
			return
		}

		if err := input.Validate(); err != nil {
			web.Error(c, http.StatusUnprocessableEntity, ErrInvalidJson.Error())
			return
		}

		id := uuid.New()

		newPessoa := domain.Pessoa{
			ID:         id.String(),
			Apelido:    input.Apelido,
			Nome:       input.Nome,
			Nascimento: input.Nascimento,
			Stack:      strings.Join(input.Stack, ","),
		}

		pessoaSaved, err := p.pessoaService.Save(c, newPessoa)
		if err != nil {
			if err.Error() == pessoa.ErrDuplicateApelido.Error() {
				web.Error(c, http.StatusUnprocessableEntity, err.Error())
				return
			}
			web.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		newPessoaResponse := PessoaResponse{
			ID:         pessoaSaved.ID,
			Apelido:    pessoaSaved.Apelido,
			Nome:       pessoaSaved.Nome,
			Nascimento: pessoaSaved.Nascimento,
			Stack:      strings.Split(pessoaSaved.Stack, ","),
		}

		c.Writer.Header().Set("Location", "/pessoas/"+newPessoaResponse.ID)

		web.Success(c, http.StatusCreated, newPessoaResponse)
	}
}

func (p *PessoaController) Count() gin.HandlerFunc {
	return func(c *gin.Context) {
		count, err := p.pessoaService.Count(c)
		if err != nil {
			web.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		web.Success(c, http.StatusOK, count)
	}
}
