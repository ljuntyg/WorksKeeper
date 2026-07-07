package entity

import "github.com/jackc/pgx/v5"

type Group struct {
	Id            int64  `db:"id"`
	ContentId     *int64 `db:"content_id"`
	SwapDirection bool   `db:"swap_direction"`
}

type GroupArguments struct {
	ContentId     *int64
	SwapDirection bool
}

func (ga *GroupArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"content_id":     ga.ContentId,
		"swap_direction": ga.SwapDirection,
	}
}
