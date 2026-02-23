package view

import "WorksKeeper/internal/entity"

type SourceView struct {
	source *entity.Source
}

func NewSourceView(source *entity.Source) *SourceView {
	return &SourceView{
		source: source,
	}
}

func (sv *SourceView) GetId() int64 {
	return sv.source.Id
}
