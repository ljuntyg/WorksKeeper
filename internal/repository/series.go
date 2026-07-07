package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SeriesRepository struct {
	pgxPool *pgxpool.Pool
}

func (sr *SeriesRepository) init(pgxPool *pgxpool.Pool) {
	sr.pgxPool = pgxPool
}

func (sr *SeriesRepository) GetSeries(id int64) (entity.Series, error) {
	return selectExactlyOneFromTableWhere[entity.Series](context.Background(), sr.pgxPool, "series",
		map[string]any{"id": id}, nil, nil)
}

func (sr *SeriesRepository) InsertSeries(args *entity.SeriesArguments) (entity.Series, error) {
	return insertIntoTable[entity.Series](context.Background(), sr.pgxPool, "series", args.GetNamedArgs())
}

func (sr *SeriesRepository) InsertSeriesTx(tx pgx.Tx, args *entity.SeriesArguments) (entity.Series, error) {
	return insertIntoTable[entity.Series](context.Background(), tx, "series", args.GetNamedArgs())
}

func (sr *SeriesRepository) GetNSeries(n int) ([]entity.Series, error) {
	return selectAllFromTableLimitN[entity.Series](context.Background(), sr.pgxPool, "series", int64(n))
}

func (sr *SeriesRepository) GetSeriesByListingId(listingId int64) (entity.Series, error) {
	return selectExactlyOneFromTableWhere[entity.Series](context.Background(), sr.pgxPool, "series",
		map[string]any{"listing_id": listingId}, nil, nil)
}
