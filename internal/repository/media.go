package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Media struct {
	Id        int64 `db:"id"`
	ContentId int64 `db:"content_id"`
}

type MediaArguments struct {
	ContentId int64
}

func (ma *MediaArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"content_id": ma.ContentId,
	}
}

type MediaRepository struct {
	pgxPool *pgxpool.Pool
}

func (mr *MediaRepository) init(pgxPool *pgxpool.Pool) {
	mr.pgxPool = pgxPool
}

func (mr *MediaRepository) GetMedia(id int64) (Media, error) {
	return selectExactlyOneFromTableWhere[Media](context.Background(), mr.pgxPool, "media",
		map[string]any{"id": id}, nil, nil)
}

func (mr *MediaRepository) InsertMediaTx(tx pgx.Tx, args *MediaArguments) (Media, error) {
	return insertIntoTable[Media](context.Background(), tx, "media", args.GetNamedArgs())
}

func (mr *MediaRepository) GetMediaByContentId(contentId int64) (Media, error) {
	return selectExactlyOneFromTableWhere[Media](context.Background(), mr.pgxPool, "media",
		map[string]any{"content_id": contentId}, nil, nil)
}
