package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/template"
	"context"
)

type SeriesService struct {
	repos *repository.RepositoryCollection
}

func (s *SeriesService) Init(repos *repository.RepositoryCollection) {
	s.repos = repos
}

func (s *SeriesService) GetTemplateData(ctx context.Context, seriesId int64) *template.SeriesData {
	templateSeries := mustBuildTemplateSeriesShallow(ctx, seriesId, s.repos)
	mustAttachTemplateListings(ctx, templateSeries, s.repos)

	return &template.SeriesData{
		TemplateSeries: templateSeries,
	}
}
