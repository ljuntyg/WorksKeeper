package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Listing struct {
	Id             int64          `db:"id"`
	ParentSeriesId int64          `db:"parent_series_id"`
	Position       pgtype.Numeric `db:"position"`
	ListingType    string         `db:"listing_type"`
}

type ListingArguments struct {
	ParentSeriesId int64
	Position       pgtype.Numeric
	ListingType    string
}

func (la *ListingArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
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

func (lr *ListingRepository) GetListingsByParentSeriesIdOrderByPositionAscending(ctx context.Context, parentSeriesId int64) ([]Listing, error) {
	return selectFromTableWhere[Listing](ctx, lr.pgxPool, "listings",
		map[string]any{"parent_series_id": parentSeriesId}, nil, &orderBy{column: "position", direction: Ascending}, nil)
}

func (lr *ListingRepository) GetListingsByParentSeriesIdOrderByPositionAscendingTx(ctx context.Context, tx pgx.Tx, parentSeriesId int64) ([]Listing, error) {
	return selectFromTableWhere[Listing](ctx, tx, "listings",
		map[string]any{"parent_series_id": parentSeriesId}, nil, &orderBy{column: "position", direction: Ascending}, nil)
}

// AppendListingToSeriesTx inserts a Listing at the end of its parent Series,
// computing the next position rather than taking one.
func (lr *ListingRepository) AppendListingToSeriesTx(ctx context.Context, tx pgx.Tx, parentSeriesId int64, listingType string) (Listing, error) {
	return insertIntoTableAppendPosition[Listing](ctx, tx, "listings",
		pgx.NamedArgs{
			"parent_series_id": parentSeriesId,
			"listing_type":     listingType,
		},
		"position", "parent_series_id", parentSeriesId,
	)
}
