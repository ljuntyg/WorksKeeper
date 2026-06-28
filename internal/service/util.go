package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/repository/entity"
	"WorksKeeper/internal/template/frontend"
)

func buildTemplateWork(
	workId int64,
	canvasRepo *repository.CanvasRepository,
	groupRepo *repository.GroupRepository,
	textRepo *repository.TextRepository,
	workRepo *repository.WorkRepository,
) *frontend.TemplateWork {
	work, err := workRepo.GetWork(workId)
	if err != nil {
		panic("unexpected error getting Work")
	}

	canvas, err := canvasRepo.GetCanvasByWorkId(workId)
	if err != nil {
		panic("unexpected error getting Canvas")
	}

	rootGroups, err := groupRepo.GetGroupsByCanvasIdWhereParentIsNullOrderByIdxAscending(canvas.Id)
	if err != nil {
		panic("unexpected error getting Groups")
	}

	canvasGroups := make([]frontend.Templatable, len(rootGroups))
	for i := range rootGroups {
		canvasGroups[i] = buildTemplateGroup(&rootGroups[i], groupRepo, textRepo)
	}

	return &frontend.TemplateWork{
		Work: &work,
		TemplateCanvas: &frontend.TemplateCanvas{
			Canvas:         &canvas,
			TemplateGroups: canvasGroups,
		},
	}
}

// recursively builds a group's subtree, merging its
// child groups and texts into a single idx-ordered list.
func buildTemplateGroup(
	group *entity.Group,
	groupRepo *repository.GroupRepository,
	textRepo *repository.TextRepository,
) *frontend.TemplateGroup {
	childGroups, err := groupRepo.GetGroupsByParentIdOrderByIdxAscending(group.Id)
	if err != nil {
		panic("unexpected error getting child Groups")
	}
	texts, err := textRepo.GetTextsByGroupIdOrderByIdxAscending(group.Id)
	if err != nil {
		panic("unexpected error getting Texts")
	}

	canvasItems := make([]frontend.Templatable, 0, len(childGroups)+len(texts))
	gi, ti := 0, 0
	for gi < len(childGroups) && ti < len(texts) {
		if childGroups[gi].Idx <= texts[ti].Idx {
			canvasItems = append(canvasItems, buildTemplateGroup(&childGroups[gi], groupRepo, textRepo))
			gi++
		} else {
			canvasItems = append(canvasItems, &frontend.TemplateText{Text: &texts[ti]})
			ti++
		}
	}
	for ; gi < len(childGroups); gi++ {
		canvasItems = append(canvasItems, buildTemplateGroup(&childGroups[gi], groupRepo, textRepo))
	}
	for ; ti < len(texts); ti++ {
		canvasItems = append(canvasItems, &frontend.TemplateText{Text: &texts[ti]})
	}

	return &frontend.TemplateGroup{
		Group:        group,
		Templatables: canvasItems,
	}
}

func buildTemplateSeries(
	seriesId int64,
	seriesRepo *repository.SeriesRepository,
	workRepo *repository.WorkRepository,
) *frontend.TemplateSeries {
	series, err := seriesRepo.GetSeries(seriesId)
	if err != nil {
		panic("unexpected error getting Series")
	}

	seriesWorks, err := workRepo.GetWorksBySeriesId(series.Id)
	if err != nil {
		panic("unexpected error getting Works for Series")
	}

	seriesSeries, err := seriesRepo.GetSeriesBySeriesId(series.Id)
	if err != nil {
		panic("unexpected error getting Series for Series")
	}

	listables := make([]frontend.Listable, len(seriesWorks)+len(seriesSeries))
	for i, w := range seriesWorks {
		listables[i] = &w
	}

	for i, s := range seriesSeries {
		listables[len(seriesWorks)+i] = &s
	}

	return &frontend.TemplateSeries{
		Series:    &series,
		Listables: listables,
	}
}
