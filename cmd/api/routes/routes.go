package routes

import (
	"github.com/isadoramsouza/rinha-de-backend-go/cmd/api/handler"
	"github.com/isadoramsouza/rinha-de-backend-go/internal/pessoa"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/rueidis"

	"github.com/gin-gonic/gin"
)

type Router interface {
	MapRoutes()
}

type router struct {
	eng   *gin.Engine
	rg    *gin.RouterGroup
	db    *pgxpool.Pool
	cache rueidis.Client
}

func NewRouter(eng *gin.Engine, db *pgxpool.Pool, redis rueidis.Client) Router {
	return &router{eng: eng, db: db, cache: redis}
}

func (r *router) MapRoutes() {
	r.setGroup()
	r.buildPessoasRoutes()
}

func (r *router) setGroup() {
	r.rg = r.eng.Group("")
}

func (r *router) buildPessoasRoutes() {
	repo := pessoa.NewRepository(r.db, r.cache)
	service := pessoa.NewService(repo)
	handler := handler.NewPessoa(service)
	r.rg.GET("/pessoas/:id", handler.Get())
	r.rg.POST("/pessoas", handler.Create())
	r.rg.GET("/pessoas", handler.SearchByTerm())
	r.rg.GET("/contagem-pessoas", handler.Count())
}
