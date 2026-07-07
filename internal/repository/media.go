package repository

import (
	"WorksKeeper/internal/repository/entity"
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaRepository struct {
	pgxPool *pgxpool.Pool
}

func (mr *MediaRepository) init(pgxPool *pgxpool.Pool) {
	mr.pgxPool = pgxPool
}

func (mr *MediaRepository) GetMedia(id int64) (entity.Media, error) {
	return selectExactlyOneFromTableWhere[entity.Media](context.Background(), mr.pgxPool, "media",
		map[string]any{"id": id}, nil, nil)
}

func (mr *MediaRepository) InsertMediaTx(tx pgx.Tx, args *entity.MediaArguments) (entity.Media, error) {
	return insertIntoTable[entity.Media](context.Background(), tx, "media", args.GetNamedArgs())
}

func (mr *MediaRepository) GetMediaByContentId(contentId int64) (entity.Media, error) {
	return selectExactlyOneFromTableWhere[entity.Media](context.Background(), mr.pgxPool, "media",
		map[string]any{"content_id": contentId}, nil, nil)
}
