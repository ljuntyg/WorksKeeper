package repository

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Content struct {
	Id            int64          `db:"id"`
	WorkId        *int64         `db:"work_id"`
	ParentGroupId *int64         `db:"parent_group_id"`
	Position      pgtype.Numeric `db:"position"`
	ContentType   string         `db:"content_type"`
}

type ContentArguments struct {
	WorkId        *int64
	ParentGroupId *int64
	Position      pgtype.Numeric
	ContentType   string
}

func (ca *ContentArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"work_id":         ca.WorkId,
		"parent_group_id": ca.ParentGroupId,
		"position":        ca.Position,
		"content_type":    ca.ContentType,
	}
}

type ContentRepository struct {
	pgxPool *pgxpool.Pool
}

func (cr *ContentRepository) init(pgxPool *pgxpool.Pool) {
	cr.pgxPool = pgxPool
}

func (cr *ContentRepository) GetOneContentById(ctx context.Context, id int64) (Content, error) {
	return selectExactlyOneFromTableWhere[Content](ctx, cr.pgxPool, "contents",
		map[string]any{"id": id}, nil, nil)
}

func (cr *ContentRepository) InsertContentTx(ctx context.Context, tx pgx.Tx, args *ContentArguments) (Content, error) {
	return insertIntoTable[Content](ctx, tx, "contents", args.GetNamedArgs())
}

func (cr *ContentRepository) GetContentsByWorkIdOrderByPositionAscending(ctx context.Context, workId int64) ([]Content, error) {
	return selectFromTableWhere[Content](ctx, cr.pgxPool, "contents",
		map[string]any{"work_id": workId}, nil, &orderBy{column: "position", direction: Ascending}, nil)
}

func (cr *ContentRepository) GetContentsByWorkIdOrderByPositionAscendingTx(ctx context.Context, tx pgx.Tx, workId int64) ([]Content, error) {
	return selectFromTableWhere[Content](ctx, tx, "contents",
		map[string]any{"work_id": workId}, nil, &orderBy{column: "position", direction: Ascending}, nil)
}

func (cr *ContentRepository) GetContentsByParentGroupIdOrderByPositionAscending(ctx context.Context, parentGroupId int64) ([]Content, error) {
	return selectFromTableWhere[Content](ctx, cr.pgxPool, "contents",
		map[string]any{"parent_group_id": parentGroupId}, nil,
		&orderBy{column: "position", direction: Ascending}, nil)
}

func (cr *ContentRepository) GetContentsByParentGroupIdOrderByPositionAscendingTx(ctx context.Context, tx pgx.Tx, parentGroupId int64) ([]Content, error) {
	return selectFromTableWhere[Content](ctx, tx, "contents",
		map[string]any{"parent_group_id": parentGroupId}, nil, &orderBy{column: "position", direction: Ascending}, nil)
}

// AppendContentToWorkTx inserts a root Content at the end of its Work.
func (cr *ContentRepository) AppendContentToWorkTx(ctx context.Context, tx pgx.Tx, workId int64, contentType string) (Content, error) {
	return insertIntoTableAppendPosition[Content](ctx, tx, "contents",
		pgx.NamedArgs{
			"work_id":         workId,
			"parent_group_id": nil,
			"content_type":    contentType,
		},
		"position", "work_id", workId,
	)
}

// AppendContentToGroupTx inserts a Content at the end of its parent Group,
// computing the next position rather than taking one.
func (cr *ContentRepository) AppendContentToGroupTx(ctx context.Context, tx pgx.Tx, parentGroupId int64, contentType string) (Content, error) {
	return insertIntoTableAppendPosition[Content](ctx, tx, "contents",
		pgx.NamedArgs{
			"work_id":         nil,
			"parent_group_id": parentGroupId,
			"content_type":    contentType,
		},
		"position", "parent_group_id", parentGroupId,
	)
}

// IncreaseContentPositionTx moves the content to a position between its next
// sibling and the sibling after that (by work_id/parent_group_id, position). If the
// next sibling is the last one, it's moved to next.position + 1. Returns
// nil, nil if contentId is already at the highest position in its group.
func (cr *ContentRepository) IncreaseContentPositionTx(ctx context.Context, tx pgx.Tx, contentId int64) (*Content, error) {
	const query = `
	WITH current AS (
		SELECT work_id, parent_group_id, position
		FROM contents
		WHERE id = @content_id
	),
	next AS (
		SELECT c.position
		FROM contents c, current
		WHERE c.work_id IS NOT DISTINCT FROM current.work_id
		AND c.parent_group_id IS NOT DISTINCT FROM current.parent_group_id
		AND c.position > current.position
		ORDER BY c.position ASC
		LIMIT 1
	),
	next_next AS (
		SELECT c.position
		FROM contents c, current
		WHERE c.work_id IS NOT DISTINCT FROM current.work_id
		AND c.parent_group_id IS NOT DISTINCT FROM current.parent_group_id
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

	rows, err := tx.Query(ctx, query, pgx.NamedArgs{"content_id": contentId})
	if err != nil {
		log.Printf("IncreaseContentPositionTx error: %s", err)
		return nil, err
	}

	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[Content])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // already at the highest position
		}
		return nil, err
	}

	return &result, nil
}

// DecreaseContentPositionTx moves the content to a position between its
// previous sibling and the sibling before that (by work_id/parent_group_id,
// position). If the previous sibling is the first one, it's moved to
// prev.position - 1. Returns nil, nil if contentId is already at the
// lowest position in its group.
func (cr *ContentRepository) DecreaseContentPositionTx(ctx context.Context, tx pgx.Tx, contentId int64) (*Content, error) {
	const query = `
    WITH current AS (
		SELECT work_id, parent_group_id, position
        FROM contents
        WHERE id = @content_id
    ),
    prev AS (
        SELECT c.position
        FROM contents c, current
		WHERE c.work_id IS NOT DISTINCT FROM current.work_id
		AND c.parent_group_id IS NOT DISTINCT FROM current.parent_group_id
        AND c.position < current.position
        ORDER BY c.position DESC
        LIMIT 1
    ),
    prev_prev AS (
        SELECT c.position
        FROM contents c, current
		WHERE c.work_id IS NOT DISTINCT FROM current.work_id
		AND c.parent_group_id IS NOT DISTINCT FROM current.parent_group_id
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
	rows, err := tx.Query(ctx, query, pgx.NamedArgs{"content_id": contentId})
	if err != nil {
		log.Printf("DecreaseContentPositionTx error: %s", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[Content])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // already at the lowest position
		}
		return nil, err
	}
	return &result, nil
}

func (cr *ContentRepository) DeleteContentByIdTx(ctx context.Context, tx pgx.Tx, contentId int64) error {
	return deleteFromTableWhere(ctx, tx, "contents",
		map[string]any{"id": contentId}, nil)
}
