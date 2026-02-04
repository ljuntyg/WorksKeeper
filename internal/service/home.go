package service

import (
	"WorksKeeper/internal/frontend"
	"WorksKeeper/internal/repository"
)

type HomeService struct {
	workRepo   *repository.WorkRepository
	seriesRepo *repository.SeriesRepository
}

func (hs *HomeService) Init(workRepo *repository.WorkRepository, seriesRepo *repository.SeriesRepository) {
	hs.workRepo = workRepo
}

func (hs *HomeService) GetNListables(n int) []frontend.Listable {
	nbrWorks := n / 2
	nbrSeries := n - nbrWorks

	works := hs.workRepo.GetNWorks(nbrWorks)
	series := hs.seriesRepo.GetNSeries(nbrSeries)

	return []frontend.Listable{works, series}
}
