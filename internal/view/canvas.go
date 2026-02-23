package view

import (
	"WorksKeeper/internal/entity"
)

type CanvasView struct {
	canvas    *entity.Canvas
	rootGroup *GroupView
}

func NewCanvasView(canvas *entity.Canvas, rootGroup *GroupView) *CanvasView {
	return &CanvasView{
		canvas:    canvas,
		rootGroup: rootGroup,
	}
}

func (cv *CanvasView) GetId() int64 {
	return cv.canvas.Id
}
