package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryCollection struct {
	pgxPool        *pgxpool.Pool
	CanvasRepo     *CanvasRepository
	CaptionRepo    *CaptionRepository
	CollectionRepo *CollectionRepository
	ContentRepo    *ContentRepository
	FileRepo       *FileRepository
	FilenameRepo   *FilenameRepository
	FilenodeRepo   *FilenodeRepository
	FileserverRepo *FileserverRepository
	GroupRepo      *GroupRepository
	InstanceRepo   *InstanceRepository
	ListingRepo    *ListingRepository
	MediaRepo      *MediaRepository
	SeriesRepo     *SeriesRepository
	SourceRepo     *SourceRepository
	TextRepo       *TextRepository
	WorkRepo       *WorkRepository
}

func (rc *RepositoryCollection) Init(
	pgxPool *pgxpool.Pool,
	canvasRepo *CanvasRepository,
	captionRepo *CaptionRepository,
	collectionRepo *CollectionRepository,
	contentRepo *ContentRepository,
	fileRepo *FileRepository,
	filenameRepo *FilenameRepository,
	filenodeRepo *FilenodeRepository,
	fileserverRepo *FileserverRepository,
	groupRepo *GroupRepository,
	instanceRepo *InstanceRepository,
	listingRepo *ListingRepository,
	mediaRepo *MediaRepository,
	seriesRepo *SeriesRepository,
	sourceRepo *SourceRepository,
	textRepo *TextRepository,
	workRepo *WorkRepository,
) {
	rc.pgxPool = pgxPool

	rc.CanvasRepo = canvasRepo
	canvasRepo.init(pgxPool)

	rc.CaptionRepo = captionRepo
	captionRepo.init(pgxPool)

	rc.CollectionRepo = collectionRepo
	collectionRepo.init(pgxPool)

	rc.ContentRepo = contentRepo
	contentRepo.init(pgxPool)

	rc.FileRepo = fileRepo
	fileRepo.init(pgxPool)

	rc.FilenameRepo = filenameRepo
	filenameRepo.init(pgxPool)

	rc.FilenodeRepo = filenodeRepo
	filenodeRepo.init(pgxPool)

	rc.FileserverRepo = fileserverRepo
	fileserverRepo.init(pgxPool)

	rc.GroupRepo = groupRepo
	groupRepo.init(pgxPool)

	rc.InstanceRepo = instanceRepo
	instanceRepo.init(pgxPool)

	rc.ListingRepo = listingRepo
	listingRepo.init(pgxPool)

	rc.MediaRepo = mediaRepo
	mediaRepo.init(pgxPool)

	rc.SeriesRepo = seriesRepo
	seriesRepo.init(pgxPool)

	rc.SourceRepo = sourceRepo
	sourceRepo.init(pgxPool)

	rc.TextRepo = textRepo
	textRepo.init(pgxPool)

	rc.WorkRepo = workRepo
	workRepo.init(pgxPool)
}

func (rc *RepositoryCollection) MustBegin(ctx context.Context) pgx.Tx {
	tx, err := rc.pgxPool.Begin(ctx)
	if err != nil {
		log.Println(err)
		panic("unexpected error beginning transaction")
	}

	return tx
}

type pgxExecutor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type SortDirection string

const (
	Ascending  SortDirection = "ASC"
	Descending SortDirection = "DESC"
)

// TODO: using config correctly?
func GetPgxPool(host, dbName, username, password, port string) *pgxpool.Pool {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		username,
		password,
		host,
		port,
		dbName,
	)

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal(err)
	}

	return pool
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

// selectFromTableWhere fetches zero or more rows matching the given conditions.
func selectFromTableWhere[T any](ctx context.Context, db pgxExecutor, tableName string, equals map[string]any, null *nullFilter, order *orderBy, limit *int64) ([]T, error) {
	query, args, err := buildWhereQuery(tableName, equals, null, order, limit)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, query, args)
	if err != nil {
		log.Printf("selectFromTableWhere error: %s", err)
		return nil, err
	}

	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
}

// selectOneFromTableWhere fetches exactly one row matching the given conditions.
func selectExactlyOneFromTableWhere[T any](ctx context.Context, db pgxExecutor, tableName string, equals map[string]any, null *nullFilter, order *orderBy) (T, error) {
	var zero T
	one := int64(1)
	query, args, err := buildWhereQuery(tableName, equals, null, order, &one)
	if err != nil {
		return zero, err
	}

	rows, err := db.Query(ctx, query, args)
	if err != nil {
		log.Println(err)
		return zero, err
	}

	defer rows.Close()
	return pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[T])
}

func selectOptionalOneFromTableWhere[T any](ctx context.Context, db pgxExecutor, tableName string, equals map[string]any, null *nullFilter, order *orderBy) (*T, error) {
	one := int64(1)
	query, args, err := buildWhereQuery(tableName, equals, null, order, &one)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(ctx, query, args)
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

func insertIntoTable[T any](ctx context.Context, db pgxExecutor, tableName string, args pgx.NamedArgs) (T, error) {
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
	rows, err := db.Query(ctx, query, args)
	if err != nil {
		log.Printf("insertIntoTable error: %s", err)
		return zero, err
	}

	defer rows.Close()
	return pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
}

func selectAllFromTableLimitN[T any](ctx context.Context, db pgxExecutor, tableName string, n int64) ([]T, error) {
	query := fmt.Sprintf(`SELECT * FROM %s LIMIT @n`, pgx.Identifier{tableName}.Sanitize())
	args := pgx.NamedArgs{
		"n": n,
	}

	rows, err := db.Query(ctx, query, args)
	if err != nil {
		log.Printf("selectNFromTable error: %s", err)
		return nil, err
	}

	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
}

// updateTableWhere updates all rows matching the given conditions and returns the updated rows.
func updateTableWhere[T any](ctx context.Context, db pgxExecutor, tableName string, set map[string]any, equals map[string]any, null *nullFilter) ([]T, error) {
	if len(set) == 0 {
		return nil, fmt.Errorf("update: no fields to set")
	}
	if len(equals) == 0 && null == nil {
		// refuse to accidentally update every row in the table
		return nil, fmt.Errorf("update: no where conditions provided")
	}

	args := pgx.NamedArgs{}

	setClauses := make([]string, 0, len(set))
	for col, val := range set {
		sanitizedCol := pgx.Identifier{col}.Sanitize()
		argName := "set_" + col
		setClauses = append(setClauses, fmt.Sprintf("%s = @%s", sanitizedCol, argName))
		args[argName] = val
	}

	conditions := make([]string, 0, len(equals)+1)
	for col, val := range equals {
		sanitizedCol := pgx.Identifier{col}.Sanitize()
		argName := "where_" + col
		conditions = append(conditions, fmt.Sprintf("%s = @%s", sanitizedCol, argName))
		args[argName] = val
	}

	if null != nil {
		sanitizedCol := pgx.Identifier{null.column}.Sanitize()
		if null.isNull {
			conditions = append(conditions, fmt.Sprintf("%s IS NULL", sanitizedCol))
		} else {
			conditions = append(conditions, fmt.Sprintf("%s IS NOT NULL", sanitizedCol))
		}
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s RETURNING *",
		pgx.Identifier{tableName}.Sanitize(),
		strings.Join(setClauses, ", "),
		strings.Join(conditions, " AND "),
	)

	rows, err := db.Query(ctx, query, args)
	if err != nil {
		log.Printf("updateTableWhere error: %s", err)
		return nil, err
	}

	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
}

// updateExactlyOneTableWhere updates exactly one row and returns it, erroring if zero or more than one row matched.
func updateExactlyOneTableWhere[T any](ctx context.Context, db pgxExecutor, tableName string, set map[string]any, equals map[string]any, null *nullFilter) (T, error) {
	var zero T

	rows, err := updateTableWhere[T](ctx, db, tableName, set, equals, null)
	if err != nil {
		return zero, err
	}
	if len(rows) != 1 {
		return zero, fmt.Errorf("expected exactly one row to be updated, got %d", len(rows))
	}

	return rows[0], nil
}

// deleteFromTableWhere deletes all rows matching the given conditions.
func deleteFromTableWhere(ctx context.Context, db pgxExecutor, tableName string, equals map[string]any, null *nullFilter) error {
	if len(equals) == 0 && null == nil {
		// refuse to accidentally delete every row in the table
		return fmt.Errorf("delete: no where conditions provided")
	}

	args := pgx.NamedArgs{}
	conditions := make([]string, 0, len(equals)+1)

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

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s",
		pgx.Identifier{tableName}.Sanitize(),
		strings.Join(conditions, " AND "),
	)

	if _, err := db.Exec(ctx, query, args); err != nil {
		log.Printf("deleteFromTableWhere error: %s", err)
		return err
	}

	return nil
}

// insertIntoTableAppendPosition inserts a row into tableName, computing
// positionColumn as (max existing position within the group + 1), or 1 if
// the group is currently empty. groupColumn/groupValue scope what counts as
// "the group" (e.g. parent_series_id, parent_group_id). args must not
// already contain positionColumn.
func insertIntoTableAppendPosition[T any](ctx context.Context, db pgxExecutor, tableName string, args pgx.NamedArgs, positionColumn, groupColumn string, groupValue any) (T, error) {
	var zero T
	if _, exists := args[positionColumn]; exists {
		return zero, fmt.Errorf("insertIntoTableAppendPosition: args must not set %q directly", positionColumn)
	}

	sanitizedTable := pgx.Identifier{tableName}.Sanitize()
	sanitizedPosCol := pgx.Identifier{positionColumn}.Sanitize()
	sanitizedGroupCol := pgx.Identifier{groupColumn}.Sanitize()

	columns := make([]string, 0, len(args)+1)
	placeholders := make([]string, 0, len(args)+1)

	for col := range args {
		columns = append(columns, pgx.Identifier{col}.Sanitize())
		placeholders = append(placeholders, "@"+col)
	}

	columns = append(columns, sanitizedPosCol)
	placeholders = append(placeholders, fmt.Sprintf(
		"COALESCE((SELECT MAX(%s) + 1 FROM %s WHERE %s = @__group_value), 1)",
		sanitizedPosCol, sanitizedTable, sanitizedGroupCol,
	))

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING *",
		sanitizedTable,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	args["__group_value"] = groupValue

	rows, err := db.Query(ctx, query, args)
	if err != nil {
		log.Printf("insertIntoTableAppendPosition error: %s", err)
		return zero, err
	}

	defer rows.Close()

	return pgx.CollectOneRow(rows, pgx.RowToStructByName[T])
}
