package entity

type Group struct {
	Id            int64  `db:"id"`
	ParentId      *int64 `db:"parent_id"`
	CanvasId      int64  `db:"canvas_id"`
	Idx           int32  `db:"idx"`
	SwapDirection bool   `db:"swap_direction"`
}

func (g *Group) GetName(prefix string) string {
	return prefix + "group"
}
