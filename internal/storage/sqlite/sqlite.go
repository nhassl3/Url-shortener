package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"urlshortner.com/m/internal/storage"
	// Optional
	//_ "github.com/mattn/go-sqlite3" // init sqlite3 driver (27 line code)
)

const (
	opNew       string = "storage.sqlite.New"
	opSaveURL   string = "storage.sqlite.SaveURL"
	opDeleteURL string = "storage.sqlite.DeleteURL"
	opGetURL    string = "storage.sqlite.GetURL"
)

type Storage struct {
	db *sql.DB
}

// New
// Create database if it not exists
func New(storagePath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", opNew, err)
	}

	stmt, err := db.Prepare(
		`CREATE TABLE IF NOT EXISTS url(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    alias TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
)`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opNew, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", opNew, err)
	}

	return &Storage{db: db}, nil
}

// SaveURL
// return id or error
func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", opSaveURL, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", opSaveURL, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", opSaveURL, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", opSaveURL, err)
	}

	return id, nil
}

// GetURL
// returned string when url is fined, but zero value with error
func (s *Storage) GetURL(alias string) (string, error) {
	stmt, err := s.db.Prepare("SELECT url FROM  url WHERE alias=?")
	if err != nil {
		return "", fmt.Errorf("%s: %w", opGetURL, err)
	}

	var resURL string
	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", opGetURL, err)
	}

	return resURL, nil
}

// DeleteURL
// return error when alias not found or something went wrong in process of deleting
func (s *Storage) DeleteURL(alias string) error {
	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias=?")
	if err != nil {
		return fmt.Errorf("%s: %w", opDeleteURL, err)
	}

	result, err := stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", opDeleteURL, err)
	}

	rowsAffected, err := result.RowsAffected() // Получаем количество затронутых строк
	if err != nil {
		return fmt.Errorf("%s: %w", opDeleteURL, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", opDeleteURL, storage.ErrAliasNoExists)
	}

	return nil
}
