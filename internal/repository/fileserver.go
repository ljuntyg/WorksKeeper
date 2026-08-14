package repository

import (
	"context"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// A disk path holds two directories. Only the one holding the stored files is
// served, so nothing that is still being written can be requested. Keeping both
// under one root keeps them on one file system, which is what lets a staged
// file be renamed into place instead of copied there.
const (
	fileserverObjectsDir = "objects"
	fileserverTempDir    = "tmp"
)

type Fileserver struct {
	Id       int64  `db:"id"`
	Scheme   string `db:"scheme"`
	Host     string `db:"host"`
	Port     int32  `db:"port"`
	DiskPath string `db:"disk_path"`
	UrlPath  string `db:"url_path"`
}

// ObjectsPath is where the stored files are, and the only directory served
// under UrlPath. The path of a Filenode is relative to it.
func (f *Fileserver) ObjectsPath() string {
	return filepath.Join(f.DiskPath, fileserverObjectsDir)
}

// TempPath is where an upload is written while it is still incomplete. Nothing
// in it is reachable over HTTP, and nothing in it is worth keeping.
func (f *Fileserver) TempPath() string {
	return filepath.Join(f.DiskPath, fileserverTempDir)
}

type FileserverArguments struct {
	Scheme   string
	Host     string
	Port     int32
	DiskPath string
	UrlPath  string
}

func (fa *FileserverArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"scheme":    fa.Scheme,
		"host":      fa.Host,
		"port":      fa.Port,
		"disk_path": fa.DiskPath,
		"url_path":  fa.UrlPath,
	}
}

type FileserverRepository struct {
	pgxPool *pgxpool.Pool
}

func (fr *FileserverRepository) init(pgxPool *pgxpool.Pool) {
	fr.pgxPool = pgxPool
}

func (fr *FileserverRepository) GetFileserver(id int64) (Fileserver, error) {
	return selectExactlyOneFromTableWhere[Fileserver](context.Background(), fr.pgxPool, "fileservers",
		map[string]any{"id": id}, nil, nil)
}

// GetFileserverOrderByIdAscending returns the Fileserver uploads are written
// to. Fileservers are seeded rather than created at runtime, and there is one
// of them; ordering by id keeps the choice stable if that ever changes.
func (fr *FileserverRepository) GetFileserverOrderByIdAscending() (Fileserver, error) {
	return selectExactlyOneFromTableWhere[Fileserver](context.Background(), fr.pgxPool, "fileservers",
		nil, nil, &orderBy{column: "id", direction: Ascending})
}
