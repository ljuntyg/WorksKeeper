package repository

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Instance struct {
	Id        int64     `db:"id"`
	Scheme    string    `db:"scheme"`
	Host      string    `db:"host"`
	Port      int32     `db:"port"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
}

// Authority leaves the port off when it is the default one for the scheme.
func (i *Instance) Authority() string {
	port := int(i.Port)

	if (i.Scheme == "http" && port == 80) ||
		(i.Scheme == "https" && port == 443) {
		return i.Host
	}

	return net.JoinHostPort(i.Host, strconv.Itoa(port))
}

type InstanceArguments struct {
	Scheme string
	Host   string
	Port   int32
	Title  string
}

func (ia *InstanceArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"scheme": ia.Scheme,
		"host":   ia.Host,
		"port":   ia.Port,
		"title":  ia.Title,
	}
}

type InstanceRepository struct {
	pgxPool *pgxpool.Pool
}

func (ir *InstanceRepository) init(pgxPool *pgxpool.Pool) {
	ir.pgxPool = pgxPool
}

func (ir *InstanceRepository) GetOneInstanceById(ctx context.Context, id int64) (Instance, error) {
	return selectExactlyOneFromTableWhere[Instance](ctx, ir.pgxPool, "instances",
		map[string]any{"id": id}, nil, nil)
}

func (ir *InstanceRepository) GetOptionalInstanceByHostAndPort(ctx context.Context, host string, port int32) (*Instance, error) {
	return selectOptionalOneFromTableWhere[Instance](ctx, ir.pgxPool, "instances",
		map[string]any{"host": host, "port": port}, nil, nil)
}

// GetOptionalInstanceOrderByIdAscending tells an empty instances table apart
// from one already holding an Instance other than the configured one.
func (ir *InstanceRepository) GetOptionalInstanceOrderByIdAscending(ctx context.Context) (*Instance, error) {
	return selectOptionalOneFromTableWhere[Instance](ctx, ir.pgxPool, "instances",
		nil, nil, &orderBy{column: "id", direction: Ascending})
}

func (ir *InstanceRepository) InsertInstanceTx(ctx context.Context, tx pgx.Tx, args *InstanceArguments) (Instance, error) {
	return insertIntoTable[Instance](ctx, tx, "instances", args.GetNamedArgs())
}
