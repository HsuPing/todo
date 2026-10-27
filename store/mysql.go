package store

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql"
)

const todoColumns = "id, title, done, created_at"

type MySQLStore struct {
	db *sql.DB
}

var _ Store = (*MySQLStore)(nil)

func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

func (m *MySQLStore) List(ctx context.Context, done *bool) ([]Todo, error) {

	var rows *sql.Rows
	var err error

	if done == nil {
		rows, err = m.db.QueryContext(ctx, "SELECT "+todoColumns+" FROM todos ORDER BY id")
	} else {
		rows, err = m.db.QueryContext(ctx, "SELECT "+todoColumns+" FROM todos WHERE done = ? ORDER BY id", *done)
	}

	if err != nil {
		return []Todo{}, err
	}

	defer rows.Close()

	todos := make([]Todo, 0)

	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return []Todo{}, err
		}

		todos = append(todos, t)
	}

	if err := rows.Err(); err != nil {
		return []Todo{}, err
	}

	return todos, nil
}

func (m *MySQLStore) Get(ctx context.Context, id int) (Todo, error) {
	row := m.db.QueryRowContext(ctx, "SELECT "+todoColumns+" FROM todos WHERE id = ?", id)
	t := Todo{}

	if err := row.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Todo{}, ErrNotFound
		}
		return Todo{}, err
	}

	return t, nil
}

func (m *MySQLStore) Create(ctx context.Context, title string) (Todo, error) {
	t, err := normalizeTitle(title)
	if err != nil {
		return Todo{}, err
	}

	result, err := m.db.ExecContext(ctx, "INSERT INTO todos (title, created_at) VALUES (?, NOW())", t)
	if err != nil {
		return Todo{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Todo{}, err
	}

	return m.Get(ctx, int(id))
}

func (m *MySQLStore) Update(ctx context.Context, id int, title *string, done *bool) (Todo, error) {

	var t string
	if title != nil {
		nt, err := normalizeTitle(*title)
		t = nt
		if err != nil {
			return Todo{}, err
		}
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return Todo{}, err
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, "SELECT "+todoColumns+" FROM todos WHERE id = ? FOR UPDATE", id)
	todo := Todo{}

	if err := row.Scan(&todo.ID, &todo.Title, &todo.Done, &todo.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Todo{}, ErrNotFound
		}
		return Todo{}, err
	}

	if done != nil {
		todo.Done = *done
	}

	if title != nil {
		todo.Title = t
	}

	_, err = tx.ExecContext(ctx, "UPDATE todos SET title = ?, done = ? WHERE id = ?", todo.Title, todo.Done, id)
	if err != nil {
		return Todo{}, err
	}

	if err := tx.Commit(); err != nil {
		return Todo{}, err
	}

	return todo, nil
}

func (m *MySQLStore) Delete(ctx context.Context, id int) error {
	result, err := m.db.ExecContext(ctx, "DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return ErrNotFound
	}

	return nil
}
