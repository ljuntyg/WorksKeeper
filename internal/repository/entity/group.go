package entity

import "github.com/jackc/pgx/v5"

type Group struct {
	Id            int64  `db:"id"`
	CanvasId      *int64 `db:"canvas_id"`
	ContentId     *int64 `db:"content_id"`
	SwapDirection bool   `db:"swap_direction"`
}

type GroupArguments struct {
	CanvasId      *int64
	ContentId     *int64
	SwapDirection bool
}

func (ga *GroupArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"canvas_id":      ga.CanvasId,
		"content_id":     ga.ContentId,
		"swap_direction": ga.SwapDirection,
	}
}
