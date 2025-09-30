package postgres

import (
	"admin_history/config"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"regexp"
	"strings"
	"time"
	_ "time"
)

type Repository struct {
	ctx context.Context
	log *zap.Logger
	cfg *config.Config
	DB  *pgxpool.Pool
}

func NewRepository(log *zap.Logger, cfg *config.Config, ctx context.Context) (*Repository, error) {
	return &Repository{
		ctx: ctx,
		log: log,
		cfg: cfg,
	}, nil
}

func (r *Repository) OnStart(_ context.Context) error {
	connectionUrl := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", r.cfg.Postgres.Host, r.cfg.Postgres.Port, r.cfg.Postgres.User, r.cfg.Postgres.Password, r.cfg.Postgres.DBName, r.cfg.Postgres.SSLMode)

	r.log.Info(connectionUrl)
	pool, err := pgxpool.New(r.ctx, connectionUrl)
	if err != nil {
		return err
	}
	r.DB = pool
	return nil
}

func (r *Repository) OnStop(_ context.Context) error {
	r.DB.Close()
	return nil
}

var re = regexp.MustCompile(`\$(\d+)`)

func quote(val any) string {
	switch v := val.(type) {
	case nil:
		return "NULL"
	case string:
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	case uuid.UUID:
		return "'" + v.String() + "'::uuid"
	case time.Time:
		return "'" + v.Format(time.RFC3339) + "'::timestamptz"
	default:
		return fmt.Sprint(v)
	}
}
func valOrNil[T any](p *T) any {
	if p == nil {
		return nil // в БД пойдёт NULL
	}
	return *p
}

var _ InterfaceRepo = (*Repository)(nil)
