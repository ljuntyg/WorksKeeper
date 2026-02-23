package view

import "WorksKeeper/internal/entity"

type TextView struct {
	text *entity.Text
}

func NewTextView(text *entity.Text) *TextView {
	return &TextView{
		text: text,
	}
}

func (tv *TextView) GetId() int64 {
	return tv.text.Id
}
