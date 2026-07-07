package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/template"
)

type SeriesService struct {
	repos *repository.RepositoryCollection
}

func (s *SeriesService) Init(repos *repository.RepositoryCollection) {
	s.repos = repos
}

func (s *SeriesService) GetTemplateData(seriesId int64) *template.SeriesData {
	templateSeries := mustBuildTemplateSeriesShallow(seriesId, s.repos)
	mustAttachTemplateListings(templateSeries, s.repos)

	return &template.SeriesData{
		TemplateSeries: templateSeries,
	}
}
