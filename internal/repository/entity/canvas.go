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

type CanvasArguments struct {
	WorkId   int64
	LastEdit time.Time
}

func (c *Canvas) GetName(prefix string) string {
	return prefix + "canvas"
}

func (ca *CanvasArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"work_id":   ca.WorkId,
		"last_edit": ca.LastEdit,
	}
}
