package repository

import (
	"context"
	"database/sql"
	"fmt"

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
		return fmt.Errorf("failed to find migrations: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrations up: %w", err)
	}

	return nil
}

type DBRepository struct {
	DB  *sql.DB
	DSN string
}

func NewDBRepository(DSN string) (*DBRepository, error) {
	err := applyMigrations(DSN)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to conect to db: %w", err)
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

func (repo *DBRepository) Save(ctx context.Context, id string, url string, userID string, isDeleted bool) (string, error) {
	var returnedShort string
	err := repo.DB.QueryRowContext(
		ctx,
		`INSERT INTO urls(short, original, user_id, is_deleted) VALUES ($1, $2, $3, $4)
		ON CONFLICT(original) DO UPDATE 
		SET original = excluded.original RETURNING short`,
		id,
		url,
		userID,
		isDeleted,
	).Scan(&returnedShort)

	if err != nil {
		return "", fmt.Errorf("failed to save url to db: %w", err)
	}
	return returnedShort, nil
}

func (repo *DBRepository) Get(ctx context.Context, id string) (entities.URLRecord, bool) {
	var url entities.URLRecord

	row := repo.DB.QueryRowContext(ctx, "SELECT original, short, is_deleted FROM urls WHERE short=$1", id)
	err := row.Scan(&url.OriginalURL, &url.Short, &url.IsDeleted)

	if err != nil {
		return entities.URLRecord{}, false
	}

	return url, true
}

func (repo *DBRepository) BatchSave(ctx context.Context, records []entities.URLRecord, userID string) error {
	tx, err := repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	for _, r := range records {
		_, err := tx.ExecContext(ctx, "INSERT INTO urls(short, original, user_id) VALUES ($1, $2, $3)", r.Short, r.OriginalURL, userID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert url while batch: %w", err)
		}
	}

	return tx.Commit()
}

func (repo *DBRepository) GetUserShortens(ctx context.Context, userID string) ([]entities.URLRecord, error) {
	rows, err := repo.DB.QueryContext(ctx, "SELECT short, original FROM urls WHERE user_id=$1 AND is_deleted=false", userID)

	if err != nil {
		return []entities.URLRecord{}, err
	}

	defer rows.Close()

	var result []entities.URLRecord

	for rows.Next() {
		var r entities.URLRecord
		if err := rows.Scan(&r.Short, &r.OriginalURL); err != nil {
			return []entities.URLRecord{}, fmt.Errorf("failed to parse row while getting all shortens: %w", err)
		}
		result = append(result, r)
	}

	if err := rows.Err(); err != nil {
		return []entities.URLRecord{}, fmt.Errorf("failed to get all shortens from db: %w", err)
	}

	return result, nil
}

func (repo *DBRepository) MarkDelete(ctx context.Context, shorts []string, userID string) error {
	tx, err := repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = repo.DB.ExecContext(
		ctx,
		`UPDATE urls
		SET is_deleted = TRUE
		WHERE user_id = $1 AND short=ANY($2)`,
		userID,
		shorts,
	)

	if err != nil {
		return fmt.Errorf("failed to mark delete in db: %w", err)
	}

	return tx.Commit()
}
