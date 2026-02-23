package view

import "WorksKeeper/internal/entity"

type MediaView struct {
	media   *entity.Media
	sources []*SourceView
	caption *CaptionView
}

func NewMediaView(media *entity.Media, sources []*SourceView, captionView *CaptionView) *MediaView {
	return &MediaView{
		media:   media,
		sources: sources,
		caption: captionView,
	}
}

func (mv *MediaView) GetId() int64 {
	return mv.media.Id
}
