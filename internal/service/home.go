package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/template"
)

type HomeService struct {
	repos *repository.RepositoryCollection
}

func (hs *HomeService) Init(repos *repository.RepositoryCollection) {
	hs.repos = repos
}

func (hs *HomeService) GetTemplateData() template.Executable {
	templateCollection := mustBuildTemplateCollectionShallow(1, hs.repos)
	templateSeries := mustAttachTemplateSeries(templateCollection, hs.repos)
	mustAttachTemplateListings(templateSeries, hs.repos)

	return &template.HomeData{
		TemplateCollection: templateCollection,
		HasMore:            false,
	}
}

/* func (hs *HomeService) getNListables(n int) []frontend.Listable {
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
*/
