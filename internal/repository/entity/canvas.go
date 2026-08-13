package entity

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type Canvas struct {
	Id       int64     `db:"id"`
	WorkId   int64     `db:"work_id"`
	LastEdit time.Time `db:"last_edit"`
}

// LastEdit is omitted so the column's DEFAULT NOW() applies on insert.
type CanvasArguments struct {
	WorkId int64
}

func (ca *CanvasArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"work_id": ca.WorkId,
	}
}
