package infra

import (
	"database/sql"

	"github.com/arnald/forum/internal/bootstrap"
	"github.com/arnald/forum/internal/config"
	"github.com/arnald/forum/internal/infra/http"
	"github.com/arnald/forum/internal/infra/storage/sqlite"
)

type Services struct {
	Repositories *sqlite.Repositories
	Server       *http.Server
}

func NewInfraProviders(db *sql.DB) Services {
	return Services{
		Repositories: sqlite.NewRepositories(db),
	}
}

func NewHTTPServer(cfg *config.ServerConfig, app *bootstrap.App) *http.Server {
	return http.NewServer(cfg, app)
}
