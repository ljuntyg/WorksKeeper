package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ListingRepository struct {
	pgxPool *pgxpool.Pool
}

func (lr *ListingRepository) init(pgxPool *pgxpool.Pool) {
	lr.pgxPool = pgxPool
}

func (lr *ListingRepository) InsertListingTx(tx pgx.Tx, args *entity.ListingArguments) (entity.Listing, error) {
	return insertIntoTable[entity.Listing](context.Background(), tx, "listings", args.GetNamedArgs())
}

func (lr *ListingRepository) GetListingsByParentSeriesIdOrderByPositionAscending(parentSeriesId int64) ([]entity.Listing, error) {
	return selectFromTableWhere[entity.Listing](context.Background(), lr.pgxPool, "listings",
		map[string]any{"parent_series_id": parentSeriesId}, nil, &orderBy{column: "position", direction: Ascending}, nil)
}

func (lr *ListingRepository) GetListingsByParentSeriesIdOrderByPositionAscendingTx(tx pgx.Tx, parentSeriesId int64) ([]entity.Listing, error) {
	return selectFromTableWhere[entity.Listing](context.Background(), tx, "listings",
		map[string]any{"parent_series_id": parentSeriesId}, nil, &orderBy{column: "position", direction: Ascending}, nil)
}

func (lr *ListingRepository) InsertListingAppendTx(tx pgx.Tx, parentSeriesId int64, listingType string) (entity.Listing, error) {
	return insertIntoTableAppendPosition[entity.Listing](context.Background(), tx, "listings",
		pgx.NamedArgs{
			"parent_series_id": parentSeriesId,
			"listing_type":     listingType,
		},
		"position", "parent_series_id", parentSeriesId,
	)
}
