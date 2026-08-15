package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Filename struct {
	Id     int64  `db:"id"`
	FileId int64  `db:"file_id"`
	Name   string `db:"name"`
}

type FilenameArguments struct {
	FileId int64
	Name   string
}

func (fa *FilenameArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"file_id": fa.FileId,
		"name":    fa.Name,
	}
}

type FilenameRepository struct {
	pgxPool *pgxpool.Pool
}

func (fr *FilenameRepository) init(pgxPool *pgxpool.Pool) {
	fr.pgxPool = pgxPool
}

func (fr *FilenameRepository) GetOneFilenameById(ctx context.Context, id int64) (Filename, error) {
	return selectExactlyOneFromTableWhere[Filename](ctx, fr.pgxPool, "filenames",
		map[string]any{"id": id}, nil, nil)
}

// GetOptionalFilenameByFileIdAndNameTx looks a Filename up by the name it
// belongs to a File under, so that uploading the same file under a name it
// already has reuses the Filename instead of failing its unique constraint.
func (fr *FilenameRepository) GetOptionalFilenameByFileIdAndNameTx(ctx context.Context, tx pgx.Tx, fileId int64, name string) (*Filename, error) {
	return selectOptionalOneFromTableWhere[Filename](ctx, tx, "filenames",
		map[string]any{"file_id": fileId, "name": name}, nil, nil)
}

func (fr *FilenameRepository) InsertFilenameTx(ctx context.Context, tx pgx.Tx, args *FilenameArguments) (Filename, error) {
	return insertIntoTable[Filename](ctx, tx, "filenames", args.GetNamedArgs())
}
