package entity

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Content struct {
	Id            int64          `db:"id"`
	ParentGroupId int64          `db:"parent_group_id"`
	Position      pgtype.Numeric `db:"position"`
	ContentType   string         `db:"content_type"`
}

type ContentArguments struct {
	ParentGroupId int64
	Position      pgtype.Numeric
	ContentType   string
}

func (ca *ContentArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"parent_group_id": ca.ParentGroupId,
		"position":        ca.Position,
		"content_type":    ca.ContentType,
	}
}
