package repository

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Listing struct {
	Id             int64          `db:"id"`
	InstanceId     *int64         `db:"instance_id"`
	ParentSeriesId *int64         `db:"parent_series_id"`
	Position       pgtype.Numeric `db:"position"`
	ListingType    string         `db:"listing_type"`
}

type ListingArguments struct {
	InstanceId     *int64
	ParentSeriesId *int64
	Position       pgtype.Numeric
	ListingType    string
}

func (la *ListingArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"instance_id":      la.InstanceId,
		"parent_series_id": la.ParentSeriesId,
		"position":         la.Position,
		"listing_type":     la.ListingType,
	}
}

type ListingRepository struct {
	pgxPool *pgxpool.Pool
}

func (lr *ListingRepository) init(pgxPool *pgxpool.Pool) {
	lr.pgxPool = pgxPool
}

func (lr *ListingRepository) InsertListingTx(ctx context.Context, tx pgx.Tx, args *ListingArguments) (Listing, error) {
	return insertIntoTable[Listing](ctx, tx, "listings", args.GetNamedArgs())
}

func (lr *ListingRepository) GetListingsByInstanceIdOrderByPositionAscending(ctx context.Context, instanceId int64) ([]Listing, error) {
	return selectFromTableWhere[Listing](ctx, lr.pgxPool, "listings",
		map[string]any{"instance_id": instanceId}, nil, &orderBy{column: "position", direction: Ascending}, nil)
}

func (lr *ListingRepository) GetListingsByInstanceIdOrderByPositionAscendingTx(ctx context.Context, tx pgx.Tx, instanceId int64) ([]Listing, error) {
	return selectFromTableWhere[Listing](ctx, tx, "listings",
		map[string]any{"instance_id": instanceId}, nil, &orderBy{column: "position", direction: Ascending}, nil)
}

func (lr *ListingRepository) GetListingsByParentSeriesIdOrderByPositionAscending(ctx context.Context, parentSeriesId int64) ([]Listing, error) {
	return selectFromTableWhere[Listing](ctx, lr.pgxPool, "listings",
		map[string]any{"parent_series_id": parentSeriesId}, nil, &orderBy{column: "position", direction: Ascending}, nil)
}

func (lr *ListingRepository) GetListingsByParentSeriesIdOrderByPositionAscendingTx(ctx context.Context, tx pgx.Tx, parentSeriesId int64) ([]Listing, error) {
	return selectFromTableWhere[Listing](ctx, tx, "listings",
		map[string]any{"parent_series_id": parentSeriesId}, nil, &orderBy{column: "position", direction: Ascending}, nil)
}

// AppendListingToInstanceTx inserts a root Listing at the end of its Instance.
func (lr *ListingRepository) AppendListingToInstanceTx(ctx context.Context, tx pgx.Tx, instanceId int64, listingType string) (Listing, error) {
	return insertIntoTableAppendPosition[Listing](ctx, tx, "listings",
		pgx.NamedArgs{
			"instance_id":      instanceId,
			"parent_series_id": nil,
			"listing_type":     listingType,
		},
		"position", "instance_id", instanceId,
	)
}

// AppendListingToSeriesTx inserts a Listing at the end of its parent Series,
// computing the next position rather than taking one.
func (lr *ListingRepository) AppendListingToSeriesTx(ctx context.Context, tx pgx.Tx, parentSeriesId int64, listingType string) (Listing, error) {
	return insertIntoTableAppendPosition[Listing](ctx, tx, "listings",
		pgx.NamedArgs{
			"instance_id":      nil,
			"parent_series_id": parentSeriesId,
			"listing_type":     listingType,
		},
		"position", "parent_series_id", parentSeriesId,
	)
}

func (lr *ListingRepository) IncreaseListingPositionTx(ctx context.Context, tx pgx.Tx, listingId int64) (*Listing, error) {
	const query = `
	WITH current AS (
		SELECT instance_id, parent_series_id, position
		FROM listings
		WHERE id = @listing_id
	),
	next AS (
		SELECT l.position
		FROM listings l, current
		WHERE l.instance_id IS NOT DISTINCT FROM current.instance_id
		AND l.parent_series_id IS NOT DISTINCT FROM current.parent_series_id
		AND l.position > current.position
		ORDER BY l.position ASC
		LIMIT 1
	),
	next_next AS (
		SELECT l.position
		FROM listings l, current
		WHERE l.instance_id IS NOT DISTINCT FROM current.instance_id
		AND l.parent_series_id IS NOT DISTINCT FROM current.parent_series_id
		AND l.position > (SELECT position FROM next)
		ORDER BY l.position ASC
		LIMIT 1
	)
	UPDATE listings
	SET position = COALESCE(
		(SELECT (next.position + next_next.position) / 2 FROM next, next_next),
		(SELECT next.position + 1 FROM next)
	)
	WHERE id = @listing_id
	AND EXISTS (SELECT 1 FROM next)
	RETURNING *;
	`

	rows, err := tx.Query(ctx, query, pgx.NamedArgs{"listing_id": listingId})
	if err != nil {
		log.Printf("IncreaseListingPositionTx error: %s", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[Listing])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}

func (lr *ListingRepository) DecreaseListingPositionTx(ctx context.Context, tx pgx.Tx, listingId int64) (*Listing, error) {
	const query = `
	WITH current AS (
		SELECT instance_id, parent_series_id, position
		FROM listings
		WHERE id = @listing_id
	),
	prev AS (
		SELECT l.position
		FROM listings l, current
		WHERE l.instance_id IS NOT DISTINCT FROM current.instance_id
		AND l.parent_series_id IS NOT DISTINCT FROM current.parent_series_id
		AND l.position < current.position
		ORDER BY l.position DESC
		LIMIT 1
	),
	prev_prev AS (
		SELECT l.position
		FROM listings l, current
		WHERE l.instance_id IS NOT DISTINCT FROM current.instance_id
		AND l.parent_series_id IS NOT DISTINCT FROM current.parent_series_id
		AND l.position < (SELECT position FROM prev)
		ORDER BY l.position DESC
		LIMIT 1
	)
	UPDATE listings
	SET position = COALESCE(
		(SELECT (prev.position + prev_prev.position) / 2 FROM prev, prev_prev),
		(SELECT prev.position - 1 FROM prev)
	)
	WHERE id = @listing_id
	AND EXISTS (SELECT 1 FROM prev)
	RETURNING *;
	`

	rows, err := tx.Query(ctx, query, pgx.NamedArgs{"listing_id": listingId})
	if err != nil {
		log.Printf("DecreaseListingPositionTx error: %s", err)
		return nil, err
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[Listing])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}

func (lr *ListingRepository) DeleteListingByIdTx(ctx context.Context, tx pgx.Tx, listingId int64) error {
	return deleteFromTableWhere(ctx, tx, "listings", map[string]any{"id": listingId}, nil)
}
