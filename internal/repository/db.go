package repository

import (
	"context"
	"database/sql"

	"github.com/Oleg2210/goshortener/internal/entities"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func applyMigrations(dsn string) error {
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

type DBRepository struct {
	DB  *sql.DB
	DSN string
}

func NewDBRepository(DSN string) (*DBRepository, error) {
	error := applyMigrations(DSN)
	if error != nil {
		return nil, error
	}

	db, error := sql.Open("pgx", DSN)
	if error != nil {
		return nil, error
	}

	repo := DBRepository{
		DB:  db,
		DSN: DSN,
	}

	return &repo, nil
}
func (repo *DBRepository) Ping(ctx context.Context) bool {
	err := repo.DB.PingContext(ctx)
	return err == nil
}

func (repo *DBRepository) Save(ctx context.Context, id string, url string) (string, error) {
	var returnedShort string
	err := repo.DB.QueryRowContext(
		ctx,
		`INSERT INTO urls(short, original) VALUES ($1, $2)
		ON CONFLICT(original) DO UPDATE 
		SET original = excluded.original RETURNING short`,
		id,
		url,
	).Scan(&returnedShort)

	if err != nil {
		return "", err
	}
	return returnedShort, nil
}

func (repo *DBRepository) Get(ctx context.Context, id string) (string, bool) {
	var fullURL string

	row := repo.DB.QueryRowContext(ctx, "SELECT original FROM urls WHERE short=$1", id)
	err := row.Scan(&fullURL)

	if err != nil {
		return "", false
	}

	return fullURL, true
}

func (repo *DBRepository) BatchSave(ctx context.Context, records []entities.URLRecord) error {
	tx, err := repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, r := range records {
		_, err := tx.ExecContext(ctx, "INSERT INTO urls(short, original) VALUES ($1, $2)", r.Short, r.OriginalURL)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
