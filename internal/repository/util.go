package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SortDirection string

const (
	Ascending  SortDirection = "ASC"
	Descending SortDirection = "DESC"
)

func GetPgxPool(host, dbName, username, password, port string) *pgxpool.Pool {
	pgxPool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, dbName))
	if err != nil {
		log.Fatal(err)
	}

	return pgxPool
}

func (sd SortDirection) sanitize() (string, error) {
	switch sd {
	case Ascending, Descending:
		return string(sd), nil
	default:
		return "", fmt.Errorf("invalid sort direction: %q", sd)
	}
}

// nullFilter optionally adds "column IS NULL" / "column IS NOT NULL" to a query.
type nullFilter struct {
	column string
	isNull bool
}

// orderBy optionally adds "ORDER BY column ASC/DESC" to a query.
type orderBy struct {
	column    string
	direction SortDirection
}

// buildWhereQuery is shared by selectFromTableWhere and selectOneFromTableWhere.
func buildWhereQuery(tableName string, equals map[string]any, null *nullFilter, order *orderBy, limit *int64) (string, pgx.NamedArgs, error) {
	conditions := make([]string, 0, len(equals)+1)
	args := pgx.NamedArgs{}

	for col, val := range equals {
		sanitizedCol := pgx.Identifier{col}.Sanitize()
		conditions = append(conditions, fmt.Sprintf("%s = @%s", sanitizedCol, col))
		args[col] = val
	}

	if null != nil {
		sanitizedCol := pgx.Identifier{null.column}.Sanitize()
		if null.isNull {
			conditions = append(conditions, fmt.Sprintf("%s IS NULL", sanitizedCol))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s IS NOT NULL", sanitizedCol))
		}
	}

	query := fmt.Sprintf("SELECT * FROM %s", pgx.Identifier{tableName}.Sanitize())
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if order != nil {
		dir, err := order.direction.sanitize()
		if err != nil {
			return "", nil, err
		}
		sanitizedCol := pgx.Identifier{order.column}.Sanitize()
		query += fmt.Sprintf(" ORDER BY %s %s", sanitizedCol, dir)
	}

	if limit != nil {
		query += " LIMIT @__limit"
		args["__limit"] = *limit
	}

	return query, args, nil
}

// TODO: what context should callers pass?
// selectFromTableWhere fetches zero or more rows matching the given conditions.
func selectFromTableWhere[T any](ctx context.Context, pgxPool *pgxpool.Pool, tableName string, equals map[string]any, null *nullFilter, order *orderBy, limit *int64) ([]T, error) {
	query, args, err := buildWhereQuery(tableName, equals, null, order, limit)
	if err != nil {
		return nil, err
	}

	rows, err := pgxPool.Query(ctx, query, args)
	if err != nil {
		log.Printf("selectFromTableWhere error: %s", err)
		return nil, err
	}

	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
}

// TODO: what context should callers pass?
// selectOneFromTableWhere fetches exactly one row matching the given conditions.
func selectExactlyOneFromTableWhere[T any](ctx context.Context, pgxPool *pgxpool.Pool, tableName string, equals map[string]any, null *nullFilter, order *orderBy) (T, error) {
	var zero T
	one := int64(1)
	query, args, err := buildWhereQuery(tableName, equals, null, order, &one)
	if err != nil {
		return zero, err
	}

	rows, err := pgxPool.Query(ctx, query, args)
	if err != nil {
		log.Println(err)
		return zero, err
	}

	defer rows.Close()
	return pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[T])
}

func selectOptionalOneFromTableWhere[T any](ctx context.Context, pgxPool *pgxpool.Pool, tableName string, equals map[string]any, null *nullFilter, order *orderBy) (*T, error) {
	one := int64(1)
	query, args, err := buildWhereQuery(tableName, equals, null, order, &one)
	if err != nil {
		return nil, err
	}

	rows, err := pgxPool.Query(ctx, query, args)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &result, nil
}

// TODO: what context should callers pass?
func insertIntoTable[T any](ctx context.Context, pgxPool *pgxpool.Pool, tableName string, args pgx.NamedArgs) (T, error) {
	var zero T
	if len(args) == 0 {
		return zero, fmt.Errorf("insert: no arguments provided")
	}

	columns := make([]string, 0, len(args))
	placeholders := make([]string, 0, len(args))
	for col := range args {
		columns = append(columns, col)
		placeholders = append(placeholders, "@"+col)
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s) RETURNING *`,
		pgx.Identifier{tableName}.Sanitize(),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	rows, err := pgxPool.Query(ctx, query, args)
	if err != nil {
		log.Printf("insertIntoTable error: %s", err)
		return zero, err
	}

	defer rows.Close()
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
}

// TODO: what context should callers pass?
func selectAllFromTableLimitN[T any](ctx context.Context, pgxPool *pgxpool.Pool, tableName string, n int64) ([]T, error) {
	query := fmt.Sprintf(`SELECT * FROM %s LIMIT @n`, pgx.Identifier{tableName}.Sanitize())
	args := pgx.NamedArgs{
		"n": n,
	}

	rows, err := pgxPool.Query(ctx, query, args)
	if err != nil {
		log.Printf("selectNFromTable error: %s", err)
		return nil, err
	}

	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
}
