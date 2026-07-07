package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/template"
	"WorksKeeper/internal/template/frontend"
	"context"
	"log"
)

type CanvasService struct {
	repos *repository.RepositoryCollection
}

func (cs *CanvasService) Init(repos *repository.RepositoryCollection) {
	cs.repos = repos
}

func (cs *CanvasService) GetTemplateData(workId int64, editing bool) template.Executable {
	templateWork := mustBuildTemplateWorkShallow(workId, cs.repos)
	mustFillTemplateWork(templateWork, cs.repos)

	return &template.CanvasData{
		TemplateWork: templateWork,
		IsEditing:    editing,
	}
}

func (cs *CanvasService) MustInsertNewWorkInBaseCollection() *frontend.TemplateWork {
	collection, err := cs.repos.CollectionRepo.GetCollection(1)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Collection")
	}

	return cs.mustInsertNewWork(collection.RootSeriesId)
}

func (cs *CanvasService) MustInsertNewTextInGroup(groupId int64) *frontend.TemplateText {
	tx := cs.repos.MustBegin(context.Background())

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.Background())
			log.Println(r)
			panic("unexpected error inserting new Text; rolled back")
		}
	}()

	templateContent := mustInsertNewTemplateTextContent(tx, groupId, cs.repos)

	if err := tx.Commit(context.Background()); err != nil {
		log.Println(err)
		panic("unexpected error committing new Text")
	}

	return templateContent.TemplateGroupOrTextOrMedia.(*frontend.TemplateText)
}

func (cs *CanvasService) MustInsertNewMediaInGroup(groupId int64) *frontend.TemplateMedia {
	tx := cs.repos.MustBegin(context.Background())

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.Background())
			log.Println(r)
			panic("unexpected error inserting new Media; rolled back")
		}
	}()

	templateContent := mustInsertNewTemplateMediaContent(tx, groupId, cs.repos)

	if err := tx.Commit(context.Background()); err != nil {
		log.Println(err)
		panic("unexpected error committing new Media")
	}

	return templateContent.TemplateGroupOrTextOrMedia.(*frontend.TemplateMedia)
}

func (cs *CanvasService) MustInsertNewGroupInGroup(groupId int64) *frontend.TemplateGroup {
	tx := cs.repos.MustBegin(context.Background())

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.Background())
			log.Println(r)
			panic("unexpected error inserting new Group; rolled back")
		}
	}()

	templateContent := mustInsertNewTemplateGroupContent(tx, groupId, cs.repos)

	if err := tx.Commit(context.Background()); err != nil {
		log.Println(err)
		panic("unexpected error committing new Group")
	}

	return templateContent.TemplateGroupOrTextOrMedia.(*frontend.TemplateGroup)
}

func (cs *CanvasService) MustIncreaseContentPosition(contentId int64) {
	tx := cs.repos.MustBegin(context.Background())
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.Background())
			log.Println(r)
			panic("unexpected error increasing Content position; rolled back")
		}
	}()

	if _, err := cs.repos.ContentRepo.IncreaseContentPositionTx(tx, contentId); err != nil {
		log.Println(err)
		panic("unexpected error increasing Content position")
	}

	if err := tx.Commit(context.Background()); err != nil {
		log.Println(err)
		panic("unexpected error increasing Content position")
	}
}

func (cs *CanvasService) MustDecreaseContentPosition(contentId int64) {
	tx := cs.repos.MustBegin(context.Background())
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.Background())
			log.Println(r)
			panic("unexpected error decreasing Content position; rolled back")
		}
	}()

	if _, err := cs.repos.ContentRepo.DecreaseContentPositionTx(tx, contentId); err != nil {
		log.Println(err)
		panic("unexpected error decreasing Content position")
	}

	if err := tx.Commit(context.Background()); err != nil {
		log.Println(err)
		panic("unexpected error decreasing Content position")
	}
}

func (cs *CanvasService) mustInsertNewWork(seriesId int64) *frontend.TemplateWork {
	tx := cs.repos.MustBegin(context.Background())

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.Background())
			log.Println(r)
			panic("unexpected error inserting new Work; rolled back")
		}
	}()

	templateListing := mustInsertNewTemplateWorkListing(tx, seriesId, cs.repos)

	if err := tx.Commit(context.Background()); err != nil {
		log.Println(err)
		panic("unexpected error committing new Work")
	}

	return templateListing.TemplateWorkOrSeries.(*frontend.TemplateWork)
}
