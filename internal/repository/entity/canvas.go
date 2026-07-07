package entity

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type Canvas struct {
	Id          int64     `db:"id"`
	RootGroupId int64     `db:"root_group_id"`
	LastEdit    time.Time `db:"last_edit"`
}

type CanvasArguments struct {
	RootGroupId int64
	LastEdit    time.Time
}

func (ca *CanvasArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"root_group_id": ca.RootGroupId,
		"last_edit":     ca.LastEdit,
	}
}
