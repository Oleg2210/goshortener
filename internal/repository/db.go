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
		return "", err
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
		return err
	}

	for _, r := range records {
		_, err := tx.ExecContext(ctx, "INSERT INTO urls(short, original, user_id) VALUES ($1, $2, $3)", r.Short, r.OriginalURL, userID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (repo *DBRepository) GetUserShortens(ctx context.Context, userID string) ([]entities.URLRecord, error) {
	rows, err := repo.DB.QueryContext(ctx, "SELECT short, original FROM urls WHERE user_id=$1", userID)

	if err != nil {
		return []entities.URLRecord{}, err
	}

	defer rows.Close()

	var result []entities.URLRecord

	for rows.Next() {
		var r entities.URLRecord
		if err := rows.Scan(&r.Short, &r.OriginalURL); err != nil {
			return []entities.URLRecord{}, err
		}
		result = append(result, r)
	}

	if err := rows.Err(); err != nil {
		return []entities.URLRecord{}, err
	}

	return result, nil
}

func (repo *DBRepository) MarkDelete(ctx context.Context, short string, userID string) error {
	_, err := repo.DB.ExecContext(
		ctx,
		`UPDATE urls
        SET is_deleted = TRUE
        WHERE short = $1 AND user_id = $2`,
		short,
		userID,
	)

	return err
}
