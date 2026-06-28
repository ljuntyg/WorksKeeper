package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/template"
)

type SeriesService struct {
	seriesRepo *repository.SeriesRepository
	workRepo   *repository.WorkRepository
}

func (s *SeriesService) Init(
	seriesRepo *repository.SeriesRepository,
	workRepo *repository.WorkRepository) {
	s.seriesRepo = seriesRepo
	s.workRepo = workRepo
}

func (s *SeriesService) GetTemplateData(seriesId int64) *template.SeriesData {
	return &template.SeriesData{
		TemplateSeries: buildTemplateSeries(seriesId, s.seriesRepo, s.workRepo),
	}
}
