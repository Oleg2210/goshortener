package repository

import (
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

func (repo *DBRepository) Ping() bool {
	error := repo.DB.Ping()
	return error == nil
}

func (repo *DBRepository) Save(id string, url string) (string, error) {
	var returnedShort string
	err := repo.DB.QueryRow(
		`insert into urls(short, original) values ($1, $2)
		on conflict(original) do update 
		set original = excluded.original returning short`,
		id,
		url,
	).Scan(&returnedShort)

	if err != nil {
		return "", err
	}
	return returnedShort, nil
}

func (repo *DBRepository) Get(id string) (string, bool) {
	var fullURL string

	row := repo.DB.QueryRow("select original from urls where short=$1", id)

	if err := row.Scan(&fullURL); err == nil {
		return fullURL, true
	}

	return "", false
}

func (repo *DBRepository) BatchSave(
	records []entities.URLRecord,
) error {
	tx, err := repo.DB.Begin()
	if err != nil {
		return err
	}

	for _, r := range records {
		_, err := repo.DB.Exec("insert into urls(short, original) values ($1, $2)", r.Short, r.OriginalURL)

		if err != nil {
			tx.Rollback()
			return err
		}

	}

	return tx.Commit()
}
