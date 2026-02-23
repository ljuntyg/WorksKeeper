package view

import (
	"WorksKeeper/internal/entity"
	"WorksKeeper/internal/frontend"
)

type GroupView struct {
	group    *entity.Group
	children []frontend.Templatable
}

func NewGroupView(group *entity.Group, children []frontend.Templatable) *GroupView {
	return &GroupView{
		group:    group,
		children: children,
	}
}

func (gv *GroupView) GetId() int64 {
	return gv.group.Id
}

func (gv *GroupView) PutChild(t frontend.Templatable, idx int32) {
	gv.children[idx] = t
}
