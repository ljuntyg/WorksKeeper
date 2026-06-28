package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/template"
	"WorksKeeper/internal/template/frontend"
)

type HomeService struct {
	seriesRepo *repository.SeriesRepository
	workRepo   *repository.WorkRepository
}

func (hs *HomeService) Init(
	seriesRepo *repository.SeriesRepository,
	workRepo *repository.WorkRepository) {
	hs.seriesRepo = seriesRepo
	hs.workRepo = workRepo
}

func (hs *HomeService) GetTemplateData() *template.HomeData {
	return &template.HomeData{
		Listables: hs.getNListables(5),
		HasMore:   false,
	}
}

func (hs *HomeService) getNListables(n int) []frontend.Listable {
	nbrWorks := n / 2
	nbrSeries := n - nbrWorks

	works, err := hs.workRepo.GetNWorks(nbrWorks)
	if err != nil {
		panic("unexpected error getting Works")
	}

	series, err := hs.seriesRepo.GetNSeries(nbrSeries)
	if err != nil {
		panic("unexpected error getting Series")
	}

	listables := make([]frontend.Listable, len(works)+len(series))
	for i, w := range works {
		listables[i] = &w
	}

	for i, s := range series {
		listables[len(works)+i] = &s
	}

	return listables
}
