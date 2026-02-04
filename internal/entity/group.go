package entity

type Group struct {
	id            int64
	parentId      int64
	canvasId      int64
	idx           int32
	swapDirection bool
}
