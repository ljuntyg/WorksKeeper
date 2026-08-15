package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Source struct {
	Id         int64 `db:"id"`
	MediaId    int64 `db:"media_id"`
	FilenameId int64 `db:"filename_id"`
}

type SourceArguments struct {
	MediaId    int64
	FilenameId int64
}

func (sa *SourceArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"media_id":    sa.MediaId,
		"filename_id": sa.FilenameId,
	}
}

type SourceRepository struct {
	pgxPool *pgxpool.Pool
}

func (sr *SourceRepository) init(pgxPool *pgxpool.Pool) {
	sr.pgxPool = pgxPool
}

func (sr *SourceRepository) GetSourcesByMediaId(ctx context.Context, mediaId int64) ([]Source, error) {
	return selectFromTableWhere[Source](ctx, sr.pgxPool, "sources",
		map[string]any{"media_id": mediaId}, nil, nil, nil)
}

func (sr *SourceRepository) InsertSourceTx(ctx context.Context, tx pgx.Tx, args *SourceArguments) (Source, error) {
	return insertIntoTable[Source](ctx, tx, "sources", args.GetNamedArgs())
}
