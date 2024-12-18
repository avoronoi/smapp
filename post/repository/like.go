package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"smapp/post/model"

	"github.com/google/uuid"
)

type Like struct {
	db         *sql.DB
	entityType model.EntityType
}

func NewPostLike(db *sql.DB) *Like {
	return &Like{db: db, entityType: model.PostType}
}

func NewCommentLike(db *sql.DB) *Like {
	return &Like{db: db, entityType: model.CommentType}
}

func (l *Like) Create(ctx context.Context, entityID, authorID uuid.UUID) error {
	fail := func(err error) error {
		return fmt.Errorf("add like to db: %w", err)
	}

	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()

	id, err := uuid.NewRandom()
	if err != nil {
		return fail(err)
	}

	result, err := tx.ExecContext(
		ctx,
		"INSERT IGNORE INTO likes (id, entity_type, entity_id, author_id) VALUES (?, ?, ?, ?)",
		id[:], l.entityType, entityID[:], authorID[:],
	)
	if err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fail(err)
	}
	if rowsAffected == 0 {
		return ErrRecordExists
	}

	countID, err := uuid.NewRandom()
	if err != nil {
		return fail(err)
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO likes_count (id, entity_type, entity_id, count) VALUES (?, ?, ?, 1) ON DUPLICATE KEY UPDATE count = count + 1",
		countID[:], l.entityType, entityID[:],
	)
	if err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}

	if err = tx.Commit(); err != nil {
		return fail(changeErrIfCtxDone(ctx, err))
	}
	return nil
}

func (l *Like) GetCount(ctx context.Context, entityID uuid.UUID) (uint32, error) {
	fail := func(err error) (uint32, error) {
		return 0, fmt.Errorf("get like count from db: %w", err)
	}

	var count uint32
	err := l.db.QueryRowContext(
		ctx,
		"SELECT count FROM likes_count WHERE entity_type = ? AND entity_id = ?",
		l.entityType, entityID[:],
	).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return fail(err)
	}

	return count, nil
}
