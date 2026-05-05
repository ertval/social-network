package infra

import (
	"database/sql"

	"github.com/arnald/forum/internal/app"
	"github.com/arnald/forum/internal/config"
	"github.com/arnald/forum/internal/infra/http"
	"github.com/arnald/forum/internal/infra/logger"
	"github.com/arnald/forum/internal/infra/storage/sqlite"
	"github.com/arnald/forum/internal/infra/ws"
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

func NewHTTPServer(cfg *config.ServerConfig, db *sql.DB, logger logger.Logger, appServices app.Services, hub *ws.Hub) *http.Server {
	return http.NewServer(cfg, db, logger, appServices, hub)
}
