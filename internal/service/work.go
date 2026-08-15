package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/template"
	"context"
)

type WorkService struct {
	repos *repository.RepositoryCollection
}

func (ws *WorkService) Init(repos *repository.RepositoryCollection) {
	ws.repos = repos
}

func (ws *WorkService) GetTemplateData(ctx context.Context, workId int64) *template.WorkData {
	templateWork := mustBuildTemplateWorkShallow(ctx, workId, ws.repos)
	/* templateCanvas := mustAttachTemplateCanvas(templateWork, ws.canvasRepo)
	templateGroup := mustAttachTemplateGroup(templateCanvas, ws.groupRepo)
	mustAttachTemplateContents(templateGroup, ws.contentRepo, ws.groupRepo, ws.textRepo, ws.mediaRepo, ws.sourceRepo)
	*/
	mustFillTemplateWork(ctx, templateWork, ws.repos)

	return &template.WorkData{
		TemplateWork: templateWork,
	}
}
