package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"

	"laba9/models"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

func (s *PostgresStorage) Migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
		id    SERIAL PRIMARY KEY,
		name  TEXT    NOT NULL,
		email TEXT    NOT NULL,
		age   INTEGER NOT NULL
	)`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT ''`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

func (s *PostgresStorage) List(f UserFilter) ([]models.User, error) {
	var (
		conds []string
		args  []any
	)
	n := 1

	if f.Name != "" {
		conds = append(conds, fmt.Sprintf("name ILIKE $%d", n))
		args = append(args, "%"+f.Name+"%")
		n++
	}
	if f.MinAge > 0 {
		conds = append(conds, fmt.Sprintf("age >= $%d", n))
		args = append(args, f.MinAge)
		n++
	}
	if f.MaxAge > 0 {
		conds = append(conds, fmt.Sprintf("age <= $%d", n))
		args = append(args, f.MaxAge)
		n++
	}

	query := "SELECT id, name, email, age FROM users"
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY id"

	if f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", n)
		args = append(args, f.Limit)
		n++
	}
	if f.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", n)
		args = append(args, f.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Age); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}
	return users, nil
}

func (s *PostgresStorage) Get(id int) (models.User, error) {
	var u models.User
	err := s.db.QueryRow(
		`SELECT id, name, email, age FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Age)

	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("query user: %w", err)
	}
	return u, nil
}

func (s *PostgresStorage) Create(u models.User) (models.User, error) {
	err := s.db.QueryRow(
		`INSERT INTO users (name, email, age) VALUES ($1, $2, $3) RETURNING id`,
		u.Name, u.Email, u.Age,
	).Scan(&u.ID)
	if err != nil {
		return models.User{}, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

func (s *PostgresStorage) Update(id int, u models.User) (models.User, error) {
	res, err := s.db.Exec(
		`UPDATE users SET name = $1, email = $2, age = $3 WHERE id = $4`,
		u.Name, u.Email, u.Age, id,
	)
	if err != nil {
		return models.User{}, fmt.Errorf("update user: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return models.User{}, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return models.User{}, ErrNotFound
	}
	u.ID = id
	return u, nil
}

func (s *PostgresStorage) Delete(id int) error {
	res, err := s.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStorage) CreateWithPassword(u models.User, hash string) (models.User, error) {
	err := s.db.QueryRow(
		`INSERT INTO users (name, email, age, password_hash)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		u.Name, u.Email, u.Age, hash,
	).Scan(&u.ID)
	if err != nil {
		return models.User{}, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

func (s *PostgresStorage) FindByEmail(email string) (models.User, string, error) {
	var (
		u    models.User
		hash string
	)
	err := s.db.QueryRow(
		`SELECT id, name, email, age, password_hash
		 FROM users WHERE LOWER(email) = LOWER($1)`,
		email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Age, &hash)

	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, "", ErrNotFound
	}
	if err != nil {
		return models.User{}, "", fmt.Errorf("query user: %w", err)
	}
	return u, hash, nil
}

var _ UserStorage = (*PostgresStorage)(nil)
