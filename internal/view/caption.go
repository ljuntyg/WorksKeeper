package view

import "WorksKeeper/internal/entity"

type CaptionView struct {
	caption *entity.Caption
}

func NewCaptionView(caption *entity.Caption) *CaptionView {
	return &CaptionView{
		caption: caption,
	}
}

func (cv *CaptionView) Init(caption *entity.Caption) *CaptionView {
	cv.caption = caption
	return cv
}

func (cv *CaptionView) GetId() int64 {
	return cv.caption.Id
}
