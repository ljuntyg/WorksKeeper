package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContentRepository struct {
	pgxPool *pgxpool.Pool
}

func (cr *ContentRepository) init(pgxPool *pgxpool.Pool) {
	cr.pgxPool = pgxPool
}

func (cr *ContentRepository) GetContent(id int64) (entity.Content, error) {
	return selectExactlyOneFromTableWhere[entity.Content](context.Background(), cr.pgxPool, "contents",
		map[string]any{"id": id}, nil, nil)
}

func (cr *ContentRepository) InsertContentTx(tx pgx.Tx, args *entity.ContentArguments) (entity.Content, error) {
	return insertIntoTable[entity.Content](context.Background(), tx, "contents", args.GetNamedArgs())
}

func (cr *ContentRepository) GetContentsByParentGroupIdOrderByPositionAscending(parentGroupId int64) ([]entity.Content, error) {
	return selectFromTableWhere[entity.Content](context.Background(), cr.pgxPool, "contents",
		map[string]any{"parent_group_id": parentGroupId}, nil,
		&orderBy{column: "position", direction: Ascending}, nil)
}

func (cr *ContentRepository) GetContentsByParentGroupIdOrderByPositionAscendingTx(tx pgx.Tx, parentGroupId int64) ([]entity.Content, error) {
	return selectFromTableWhere[entity.Content](context.Background(), tx, "contents",
		map[string]any{"parent_group_id": parentGroupId}, nil, &orderBy{column: "position", direction: Ascending}, nil)
}

func (cr *ContentRepository) InsertContentAppendTx(tx pgx.Tx, parentGroupId int64, contentType string) (entity.Content, error) {
	return insertIntoTableAppendPosition[entity.Content](context.Background(), tx, "contents",
		pgx.NamedArgs{
			"parent_group_id": parentGroupId,
			"content_type":    contentType,
		},
		"position", "parent_group_id", parentGroupId,
	)
}

// IncreaseContentPositionTx moves the content to a position between its next
// sibling and the sibling after that (by parent_group_id, position). If the
// next sibling is the last one, it's moved to next.position + 1. Returns
// nil, nil if contentId is already at the highest position in its group.
func (cr *ContentRepository) IncreaseContentPositionTx(tx pgx.Tx, contentId int64) (*entity.Content, error) {
	const query = `
	WITH current AS (
		SELECT parent_group_id, position
		FROM contents
		WHERE id = @content_id
	),
	next AS (
		SELECT c.position
		FROM contents c, current
		WHERE c.parent_group_id = current.parent_group_id
		AND c.position > current.position
		ORDER BY c.position ASC
		LIMIT 1
	),
	next_next AS (
		SELECT c.position
		FROM contents c, current
		WHERE c.parent_group_id = current.parent_group_id
		AND c.position > (SELECT position FROM next)
		ORDER BY c.position ASC
		LIMIT 1
	)
	UPDATE contents
	SET position = COALESCE(
		(SELECT (next.position + next_next.position) / 2 FROM next, next_next),
		(SELECT next.position + 1 FROM next)
	)
	WHERE id = @content_id
	AND EXISTS (SELECT 1 FROM next)
	RETURNING *;
	`

	rows, err := tx.Query(context.Background(), query, pgx.NamedArgs{"content_id": contentId})
	if err != nil {
		log.Printf("IncreaseContentPositionTx error: %s", err)
		return nil, err
	}

	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[entity.Content])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // already at the highest position
		}
		return nil, err
	}

	return &result, nil
}

// DecreaseContentPositionTx moves the content to a position between its
// previous sibling and the sibling before that (by parent_group_id,
// position). If the previous sibling is the first one, it's moved to
// prev.position - 1. Returns nil, nil if contentId is already at the
// lowest position in its group.
func (cr *ContentRepository) DecreaseContentPositionTx(tx pgx.Tx, contentId int64) (*entity.Content, error) {
	const query = `
    WITH current AS (
        SELECT parent_group_id, position
        FROM contents
        WHERE id = @content_id
    ),
    prev AS (
        SELECT c.position
        FROM contents c, current
        WHERE c.parent_group_id = current.parent_group_id
        AND c.position < current.position
        ORDER BY c.position DESC
        LIMIT 1
    ),
    prev_prev AS (
        SELECT c.position
        FROM contents c, current
        WHERE c.parent_group_id = current.parent_group_id
        AND c.position < (SELECT position FROM prev)
        ORDER BY c.position DESC
        LIMIT 1
    )
    UPDATE contents
    SET position = COALESCE(
        (SELECT (prev.position + prev_prev.position) / 2 FROM prev, prev_prev),
        (SELECT prev.position - 1 FROM prev)
    )
    WHERE id = @content_id
    AND EXISTS (SELECT 1 FROM prev)
    RETURNING *;
    `
	rows, err := tx.Query(context.Background(), query, pgx.NamedArgs{"content_id": contentId})
	if err != nil {
		log.Printf("DecreaseContentPositionTx error: %s", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[entity.Content])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // already at the lowest position
		}
		return nil, err
	}
	return &result, nil
}
