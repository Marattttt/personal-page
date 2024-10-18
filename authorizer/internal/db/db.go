package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Marattttt/personal-page/authorizer/pkg/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type UserRepo struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewUserRepo(conn *sqlx.DB, logger *slog.Logger) UserRepo {
	return UserRepo{
		db:     conn,
		logger: logger,
	}
}

func (u UserRepo) Get(ctx context.Context, id int) (*models.User, error) {
	const q = `
SELECT 
	id,
	role,
	login,
	pass_hash,
FROM 
	users
WHERE 
	id = $1
`

	start := time.Now()

	u.logger.Info("Executing select by id query", slog.Int("id", id))

	rows, err := u.db.QueryxContext(ctx, q, id)
	if err != nil {
		return nil, fmt.Errorf("querying: %w", err)
	}

	var user *models.User

	for rows.Next() {
		if user != nil {
			return nil, fmt.Errorf("too many rows")
		}

		if err := rows.StructScan(&user); err != nil {
			return nil, fmt.Errorf("scanning: %w", err)
		}
	}

	u.logger.Info(
		"Finished user select by id",
		slog.Duration("timeTook", time.Now().Sub(start)),
	)

	return user, nil
}

func (u UserRepo) GetLogin(ctx context.Context, login string) (*models.User, error) {
	const q = `
SELECT 
	id,
	role,
	login,
	pass_hash,
FROM 
	users
WHERE 
	id = $1
`

	start := time.Now()

	u.logger.Info("Executing select by login query", slog.String("login", login))

	rows, err := u.db.NamedQueryContext(ctx, q, login)
	if err != nil {
		return nil, fmt.Errorf("querying: %w", err)
	}

	var user *models.User

	for rows.Next() {
		if user != nil {
			return nil, fmt.Errorf("too many rows")
		}

		if err := rows.StructScan(&user); err != nil {
			return nil, fmt.Errorf("scanning: %w", err)
		}
	}

	u.logger.Info(
		"Finished user select by login",
		slog.Duration("timeTook", time.Now().Sub(start)),
	)

	return user, nil
}

func (u UserRepo) Create(ctx context.Context, user models.User) (*models.User, error) {
	const q = `
INSERT INTO users (
	login,
	role,
	pass_hash,
VALUES (
	:login,
	:role,
	:pass_hash,
RETURNING id, login, role, pass_hash`

	start := time.Now()
	slog.Info("Starting create user transaction", slog.Any("user", user))

	tx, err := u.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}
	defer rollBack(tx, u.logger)

	rows, err := tx.NamedQuery(q, user)
	if err != nil {
		return nil, fmt.Errorf("querying: %w", err)
	}

	var inserted *models.User
	for rows.Next() {
		if inserted != nil {
			return nil, fmt.Errorf("too many rows")
		}

		if err := rows.StructScan(&inserted); err != nil {
			return nil, fmt.Errorf("scanning: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing: %w", err)
	}
	slog.Info("Committed transaction sucecssfully", slog.Duration("timeTook", time.Now().Sub(start)))

	return inserted, nil
}

func rollBack(tx *sqlx.Tx, errLogger *slog.Logger) {
	err := tx.Rollback()

	// If a tx is already commited, it returns sql.ErrTxDone
	if err != nil && !errors.Is(err, sql.ErrTxDone) {
		errLogger.Error("Could not rollback transaction", slog.String("err", err.Error()))
	}
}

func Migrate(conn *sqlx.DB, source string) error {
	driver, err := postgres.WithInstance(conn.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrations driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(source, "authenticator", driver)
	if err != nil {
		return fmt.Errorf("creating migration instance: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("appying migrations: %w", err)
	}

	return nil
}
