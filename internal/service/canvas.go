package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/repository/entity"
	"WorksKeeper/internal/template"
	"log"
	"time"
)

type CanvasService struct {
	canvasRepo  *repository.CanvasRepository
	captionRepo *repository.CaptionRepository
	groupRepo   *repository.GroupRepository
	mediaRepo   *repository.MediaRepository
	sourceRepo  *repository.SourceRepository
	textRepo    *repository.TextRepository
	workRepo    *repository.WorkRepository
}

func (cs *CanvasService) Init(
	canvasRepo *repository.CanvasRepository,
	captionRepo *repository.CaptionRepository,
	groupRepo *repository.GroupRepository,
	mediaRepo *repository.MediaRepository,
	sourceRepo *repository.SourceRepository,
	textRepo *repository.TextRepository,
	workRepo *repository.WorkRepository,
) {
	cs.canvasRepo = canvasRepo
	cs.captionRepo = captionRepo
	cs.groupRepo = groupRepo
	cs.mediaRepo = mediaRepo
	cs.sourceRepo = sourceRepo
	cs.textRepo = textRepo
	cs.workRepo = workRepo
}

func (cs *CanvasService) GetTemplateData(workId int64, editing bool) *template.CanvasData {
	return &template.CanvasData{
		TemplateWork: buildTemplateWork(
			workId,
			cs.canvasRepo,
			cs.groupRepo,
			cs.textRepo,
			cs.workRepo),
		IsEditing: editing,
	}
}

func (cs *CanvasService) GetNewWork() *entity.Work {
	work, err := cs.workRepo.GetNewWork(&entity.WorkArguments{
		Title: "Untitled Work",
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error creating new Work")
	}

	_, err = cs.canvasRepo.GetNewCanvas((&entity.CanvasArguments{
		WorkId:   work.Id,
		LastEdit: time.Now(),
	}))

	if err != nil {
		log.Println(err)
		panic("unexpected error creating new Canvas")
	}

	return &work
}
