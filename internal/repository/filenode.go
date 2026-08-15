package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Filenode struct {
	Id           int64  `db:"id"`
	FileId       int64  `db:"file_id"`
	FileserverId int64  `db:"fileserver_id"`
	Path         string `db:"path"`
}

type FilenodeArguments struct {
	FileId       int64
	FileserverId int64
	Path         string
}

func (fa *FilenodeArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"file_id":       fa.FileId,
		"fileserver_id": fa.FileserverId,
		"path":          fa.Path,
	}
}

type FilenodeRepository struct {
	pgxPool *pgxpool.Pool
}

func (fr *FilenodeRepository) init(pgxPool *pgxpool.Pool) {
	fr.pgxPool = pgxPool
}

func (fr *FilenodeRepository) GetOneFilenodeById(ctx context.Context, id int64) (Filenode, error) {
	return selectExactlyOneFromTableWhere[Filenode](ctx, fr.pgxPool, "filenodes",
		map[string]any{"id": id}, nil, nil)
}

func (fr *FilenodeRepository) GetOneFilenodeByFileIdOrderByIdAscending(ctx context.Context, fileId int64) (Filenode, error) {
	return selectExactlyOneFromTableWhere[Filenode](ctx, fr.pgxPool, "filenodes",
		map[string]any{"file_id": fileId}, nil, &orderBy{column: "id", direction: Ascending})
}

func (fr *FilenodeRepository) GetOptionalFilenodeByFileserverIdAndPathTx(ctx context.Context, tx pgx.Tx, fileserverId int64, path string) (*Filenode, error) {
	return selectOptionalOneFromTableWhere[Filenode](ctx, tx, "filenodes",
		map[string]any{"fileserver_id": fileserverId, "path": path}, nil, nil)
}

func (fr *FilenodeRepository) InsertFilenodeTx(ctx context.Context, tx pgx.Tx, args *FilenodeArguments) (Filenode, error) {
	return insertIntoTable[Filenode](ctx, tx, "filenodes", args.GetNamedArgs())
}
