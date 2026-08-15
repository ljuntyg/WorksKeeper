package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type File struct {
	Id       int64  `db:"id"`
	Size     int64  `db:"size"`
	Hash     string `db:"hash"`
	MimeType string `db:"mime_type"`
}

type FileArguments struct {
	Size     int64
	Hash     string
	MimeType string
}

func (fa *FileArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"size":      fa.Size,
		"hash":      fa.Hash,
		"mime_type": fa.MimeType,
	}
}

type FileRepository struct {
	pgxPool *pgxpool.Pool
}

func (fr *FileRepository) init(pgxPool *pgxpool.Pool) {
	fr.pgxPool = pgxPool
}

func (fr *FileRepository) GetOneFileById(ctx context.Context, id int64) (File, error) {
	return selectExactlyOneFromTableWhere[File](ctx, fr.pgxPool, "files",
		map[string]any{"id": id}, nil, nil)
}

// GetOptionalFileByHashAndSizeTx looks a File up by its contents, so that an
// upload of bytes we already store reuses the File instead of adding a copy.
func (fr *FileRepository) GetOptionalFileByHashAndSizeTx(ctx context.Context, tx pgx.Tx, hash string, size int64) (*File, error) {
	return selectOptionalOneFromTableWhere[File](ctx, tx, "files",
		map[string]any{"hash": hash, "size": size}, nil, nil)
}

func (fr *FileRepository) InsertFileTx(ctx context.Context, tx pgx.Tx, args *FileArguments) (File, error) {
	return insertIntoTable[File](ctx, tx, "files", args.GetNamedArgs())
}
