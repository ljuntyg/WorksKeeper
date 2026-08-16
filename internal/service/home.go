package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/template"
	"context"
)

type HomeService struct {
	repos    *repository.RepositoryCollection
	instance *repository.Instance
}

func (hs *HomeService) Init(repos *repository.RepositoryCollection, instance *repository.Instance) {
	hs.repos = repos
	hs.instance = instance
}

func (hs *HomeService) GetTemplateData(ctx context.Context) template.Executable {
	templateInstance := buildTemplateInstanceShallow(hs.instance)
	templateCollection := mustAttachTemplateCollection(ctx, templateInstance, hs.repos)
	templateSeries := mustAttachTemplateSeries(ctx, templateCollection, hs.repos)
	mustAttachTemplateListings(ctx, templateSeries, hs.repos)

	return &template.HomeData{
		TemplateInstance: templateInstance,
		HasMore:          false,
	}
}
