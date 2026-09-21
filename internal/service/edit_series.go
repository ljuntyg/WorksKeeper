package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/template"
	"WorksKeeper/internal/template/frontend"
	"context"
	"log"
)

type EditSeriesService struct {
	repos    *repository.RepositoryCollection
	instance *repository.Instance
}

func (ess *EditSeriesService) Init(repos *repository.RepositoryCollection, instance *repository.Instance) {
	ess.repos = repos
	ess.instance = instance
}

func (ess *EditSeriesService) MustInsertNewSeriesInInstance(ctx context.Context) *frontend.TemplateSeries {
	tx := ess.repos.MustBegin(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error inserting new Series; rolled back")
		}
	}()

	templateListing := mustInsertNewTemplateSeriesListingInInstance(ctx, tx, ess.instance.Id, ess.repos)

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error committing new Series")
	}

	return templateListing.TemplateWorkOrSeries.(*frontend.TemplateSeries)
}

func (ess *EditSeriesService) GetTemplateData(ctx context.Context, seriesId int64, editing bool) template.Executable {
	templateSeries := mustBuildTemplateSeriesShallow(ctx, seriesId, ess.repos)
	mustAttachTemplateListings(ctx, templateSeries, ess.repos)

	return &template.EditSeriesData{
		TemplateSeries: templateSeries,
		IsEditing:      editing,
	}
}

func (ess *EditSeriesService) MustSaveEditSeriesEdits(ctx context.Context, seriesId int64, title *string) {
	if title == nil {
		return
	}

	tx := ess.repos.MustBegin(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error saving Series edits; rolled back")
		}
	}()

	if _, err := ess.repos.SeriesRepo.UpdateSeriesSetTitleByIdTx(ctx, tx, seriesId, *title); err != nil {
		log.Println(err)
		panic("unexpected error saving Series title")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error saving Series edits")
	}
}

func (ess *EditSeriesService) MustInsertNewWorkInSeries(ctx context.Context, seriesId int64) {
	tx := ess.repos.MustBegin(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error inserting new Work; rolled back")
		}
	}()

	mustInsertNewTemplateWorkListing(ctx, tx, seriesId, ess.repos)

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error committing new Work")
	}
}

func (ess *EditSeriesService) MustInsertNewSeriesInSeries(ctx context.Context, seriesId int64) {
	tx := ess.repos.MustBegin(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error inserting new Series; rolled back")
		}
	}()

	mustInsertNewTemplateSeriesListing(ctx, tx, seriesId, ess.repos)

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error committing new Series")
	}
}

func (ess *EditSeriesService) MustIncreaseListingPosition(ctx context.Context, listingId int64) {
	tx := ess.repos.MustBegin(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error increasing Listing position; rolled back")
		}
	}()

	if _, err := ess.repos.ListingRepo.IncreaseListingPositionTx(ctx, tx, listingId); err != nil {
		log.Println(err)
		panic("unexpected error increasing Listing position")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error committing Listing position")
	}
}

func (ess *EditSeriesService) MustDecreaseListingPosition(ctx context.Context, listingId int64) {
	tx := ess.repos.MustBegin(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error decreasing Listing position; rolled back")
		}
	}()

	if _, err := ess.repos.ListingRepo.DecreaseListingPositionTx(ctx, tx, listingId); err != nil {
		log.Println(err)
		panic("unexpected error decreasing Listing position")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error committing Listing position")
	}
}

func (ess *EditSeriesService) MustDeleteListing(ctx context.Context, listingId int64) {
	tx := ess.repos.MustBegin(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error deleting Listing; rolled back")
		}
	}()

	if err := ess.repos.ListingRepo.DeleteListingByIdTx(ctx, tx, listingId); err != nil {
		log.Println(err)
		panic("unexpected error deleting Listing")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error committing Listing deletion")
	}
}
