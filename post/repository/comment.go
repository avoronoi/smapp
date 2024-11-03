package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type Comment struct {
	db *sql.DB
}

func (c *Comment) Close() error {
	return c.db.Close()
}

func NewComment(db *sql.DB) *Comment {
	return &Comment{db: db}
}

var ErrPostDoesNotExist = fmt.Errorf("post does not exist")

func (c *Comment) Create(ctx context.Context, post_id, author_id, body string) (string, error) {
	fail := func(err error) (string, error) {
		return "", fmt.Errorf("add comment to db: %w", err)
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	id, err := uuid.NewRandom()
	if err != nil {
		return fail(err)
	}
	post_uuid, err := uuid.Parse(post_id)
	if err != nil {
		return fail(err)
	}
	author_uuid, err := uuid.Parse(author_id)
	if err != nil {
		return fail(err)
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO comments (id, post_id, author_id, body) VALUES (?, ?, ?, ?)",
		id[:], post_uuid[:], author_uuid[:], body,
	)
	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1452 {
		return fail(fmt.Errorf("%w: %s", ErrPostDoesNotExist, post_id))
	}
	if err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}

	count_id, err := uuid.NewRandom()
	if err != nil {
		return fail(err)
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO comments_count (id, post_id, count) VALUES (?, ?, 1) ON DUPLICATE KEY UPDATE count = count + 1",
		count_id[:], post_uuid[:],
	)
	if err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}

	if err = tx.Commit(); err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}
	return id.String(), nil
}
