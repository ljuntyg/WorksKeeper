package repository

import (
	"WorksKeeper/internal/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaRepository struct {
	pgxPool *pgxpool.Pool
}

func NewMediaRepository(pgxPool *pgxpool.Pool) *MediaRepository {
	return &MediaRepository{
		pgxPool: pgxPool,
	}
}

func (mr *MediaRepository) GetMediasByGroupsOrderedByGroupIdAndIdx(groups []*entity.Group) map[int64][]*entity.Media {
	return queryMultipleMapMultiple[int64, entity.Media](
		mr.pgxPool,
		"SELECT * FROM media WHERE group_id = ANY($1) ORDER BY group_id ASC, idx ASC",
		func(m *entity.Media) int64 {
			return m.GroupId
		},
		groups)
}
