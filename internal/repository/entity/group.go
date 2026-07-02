package entity

import "github.com/jackc/pgx/v5"

type Group struct {
	Id            int64  `db:"id"`
	ParentId      *int64 `db:"parent_id"`
	CanvasId      int64  `db:"canvas_id"`
	Idx           int32  `db:"idx"`
	SwapDirection bool   `db:"swap_direction"`
}

type GroupArguments struct {
	ParentId      *int64
	CanvasId      int64
	Idx           int32
	SwapDirection bool
}

func (g *Group) GetName(prefix string) string {
	return prefix + "group"
}

func (ga *GroupArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"parent_id":      ga.ParentId,
		"canvas_id":      ga.CanvasId,
		"idx":            ga.Idx,
		"swap_direction": ga.SwapDirection,
	}
}
