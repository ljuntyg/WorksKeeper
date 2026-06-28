package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/template"
)

type WorkService struct {
	canvasRepo  *repository.CanvasRepository
	captionRepo *repository.CaptionRepository
	groupRepo   *repository.GroupRepository
	mediaRepo   *repository.MediaRepository
	sourceRepo  *repository.SourceRepository
	textRepo    *repository.TextRepository
	workRepo    *repository.WorkRepository
}

func (ws *WorkService) Init(
	canvasRepo *repository.CanvasRepository,
	captionRepo *repository.CaptionRepository,
	groupRepo *repository.GroupRepository,
	mediaRepo *repository.MediaRepository,
	sourceRepo *repository.SourceRepository,
	textRepo *repository.TextRepository,
	workRepo *repository.WorkRepository) {
	ws.canvasRepo = canvasRepo
	ws.captionRepo = captionRepo
	ws.groupRepo = groupRepo
	ws.mediaRepo = mediaRepo
	ws.sourceRepo = sourceRepo
	ws.textRepo = textRepo
	ws.workRepo = workRepo
}

func (ws *WorkService) GetTemplateData(workId int64) *template.WorkData {
	return &template.WorkData{
		TemplateWork: buildTemplateWork(
			workId,
			ws.canvasRepo,
			ws.groupRepo,
			ws.textRepo,
			ws.workRepo),
	}
}
