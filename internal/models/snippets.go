// Package models provides the data structures and database models
// used throughout the application.
package models

import (
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *sql.DB
}

func (sm *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	stmt, args, err := sq.
		Insert("snippets").
		Columns("title", "content", "created", "expires").
		Values(title,
			content,
			sq.Expr("UTC_TIMESTAMP()"),
			sq.Expr("DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY)", expires),
		).ToSql()
	if err != nil {
		return 0, err
	}

	// its fine to ignore the sql.result and replace result with _
	result, err := sm.DB.Exec(stmt, args...)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (sm *SnippetModel) Get(id int) (*Snippet, error) {
	stmt, args, err := sq.
		Select("id", "title", "content", "created", "expires").
		From("snippets").
		Where("expires > UTC_TIMESTAMP() AND id = ?", id).
		ToSql()
	if err != nil {
		return nil, err
	}

	s := &Snippet{}
	err = sm.DB.QueryRow(stmt, args...).Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return s, nil
}

func (sm *SnippetModel) Latest() ([]*Snippet, error) {
	stmt, _, err := sq.
		Select("id", "title", "content", "created", "expires").
		From("snippets").
		Where("expires > UTC_TIMESTAMP()").
		OrderBy("id DESC").
		Limit(10).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := sm.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snippets := []*Snippet{}

	for rows.Next() {
		s := &Snippet{}
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}

		snippets = append(snippets, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
