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
	work, err := cs.workRepo.InsertWork(&entity.WorkArguments{
		Title: "Untitled Work",
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error creating new Work")
	}

	canvas, err := cs.canvasRepo.InsertCanvas(&entity.CanvasArguments{
		WorkId:   work.Id,
		LastEdit: time.Now(),
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error creating new Canvas")
	}

	_, err = cs.groupRepo.InsertGroup(&entity.GroupArguments{
		ParentId:      nil,
		CanvasId:      canvas.Id,
		Idx:           0,
		SwapDirection: false,
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error creating new Group")
	}

	return &work
}

func (cs *CanvasService) GetNewGroupForGroup(workId int64, groupId int64) {
	canvas, err := cs.canvasRepo.GetCanvasByWorkId(workId)
	if err != nil {
		panic("unexpected error getting Canvas")
	}

	group, err := cs.groupRepo.GetGroup(groupId)
	if err != nil {
		panic("unexpected error getting Group")
	}

	latestGroupText, err := cs.textRepo.GetTextOrNilByGroupIdOrderByIdxDescending(groupId)
	if err != nil {
		panic("unexpected error getting latest Group Text")
	}

	latestGroupMedia, err := cs.mediaRepo.GetMediaOrNilByGroupIdOrderByIdxDescending(groupId)
	if err != nil {
		panic("unexpected error getting latest Group Media")
	}

	latestGroupGroup, err := cs.groupRepo.GetGroupOrNilByParentIdOrderByIdxDescending(groupId)
	if err != nil {
		panic("unexpected error getting latest Group Group")
	}

	maxTextIdx, maxMediaIdx, maxGroupIdx := int32(0), int32(0), int32(0)
	if latestGroupText != nil {
		maxTextIdx = int32(latestGroupText.Idx)
	}

	if latestGroupMedia != nil {
		maxMediaIdx = int32(latestGroupMedia.Idx)
	}

	if latestGroupGroup != nil {
		maxGroupIdx = int32(latestGroupGroup.Idx)
	}

	maxIdx := max(maxTextIdx, maxMediaIdx, maxGroupIdx)
	if maxIdx != 0 {
		maxIdx = maxIdx + 1
	}

	cs.groupRepo.InsertGroup(&entity.GroupArguments{
		ParentId:      &groupId,
		CanvasId:      canvas.Id,
		Idx:           maxIdx,
		SwapDirection: !group.SwapDirection,
	})
}
