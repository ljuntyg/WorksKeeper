package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/template/frontend"
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

// *
// * INSTANCE INSTANCE INSTANCE
// *
// MustGetOrInsertInstance resolves the Instance this process is configured for.
// It may only create one when no Instance exists at all, so that a mistyped host
// fails at startup instead of quietly standing up a second, empty site.
func MustGetOrInsertInstance(ctx context.Context, args *repository.InstanceArguments, repos *repository.RepositoryCollection) repository.Instance {
	configured, err := repos.InstanceRepo.GetOptionalInstanceByHostAndPort(ctx, args.Host, args.Port)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Instance")
	}

	if configured != nil {
		return *configured
	}

	existing, err := repos.InstanceRepo.GetOptionalInstanceOrderByIdAscending(ctx)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Instance")
	}

	if existing != nil {
		panic("configured Instance does not exist")
	}

	tx := repos.MustBegin(ctx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error inserting new Instance; rolled back")
		}
	}()

	templateInstance := mustInsertNewTemplateInstance(ctx, tx, args, repos)

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error committing new Instance")
	}

	return *templateInstance.Instance
}

func mustInsertNewTemplateInstance(ctx context.Context, tx pgx.Tx, args *repository.InstanceArguments, repos *repository.RepositoryCollection) *frontend.TemplateInstance {
	instance, err := repos.InstanceRepo.InsertInstanceTx(ctx, tx, args)
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Instance")
	}

	return &frontend.TemplateInstance{
		Instance:         &instance,
		TemplateListings: nil,
	}
}

// The Instance is resolved once at startup, so the caller already holds the row
// and nothing has to be fetched here.
func buildTemplateInstanceShallow(instance *repository.Instance) *frontend.TemplateInstance {
	return &frontend.TemplateInstance{
		Instance:         instance,
		TemplateListings: nil,
	}
}

func mustAttachTemplateInstanceListings(ctx context.Context, ti *frontend.TemplateInstance, repos *repository.RepositoryCollection) []*frontend.TemplateListing {
	listings, err := repos.ListingRepo.GetListingsByInstanceIdOrderByPositionAscending(ctx, ti.Instance.Id)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Listings for Instance")
	}

	templateListings := mustBuildTemplateListings(ctx, listings, repos)
	ti.TemplateListings = make([]frontend.Templatable, len(templateListings))
	for i, templateListing := range templateListings {
		ti.TemplateListings[i] = templateListing
	}
	return templateListings
}

// *
// * SERIES SERIES SERIES
// *
func mustInsertNewTemplateSeriesListing(ctx context.Context, tx pgx.Tx, parentSeriesId int64, repos *repository.RepositoryCollection) *frontend.TemplateListing {
	listing, err := repos.ListingRepo.AppendListingToSeriesTx(ctx, tx, parentSeriesId, "series")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Listing")
	}

	templateSeries := mustInsertNewNestedTemplateSeries(ctx, tx, listing.Id, repos)

	return &frontend.TemplateListing{
		Listing:              &listing,
		TemplateWorkOrSeries: templateSeries,
	}
}

func mustInsertNewNestedTemplateSeries(ctx context.Context, tx pgx.Tx, listingId int64, repos *repository.RepositoryCollection) *frontend.TemplateSeries {
	series, err := repos.SeriesRepo.InsertSeriesTx(ctx, tx, &repository.SeriesArguments{
		ListingId: listingId,
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Series")
	}

	return &frontend.TemplateSeries{
		Series:           &series,
		TemplateListings: nil, // no children required to be valid
	}
}

func mustBuildTemplateSeriesShallow(ctx context.Context, seriesId int64, repos *repository.RepositoryCollection) *frontend.TemplateSeries {
	series, err := repos.SeriesRepo.GetOneSeriesById(ctx, seriesId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Series")
	}

	return &frontend.TemplateSeries{
		Series:           &series,
		TemplateListings: nil,
	}
}

func mustFillTemplateSeries(ctx context.Context, ts *frontend.TemplateSeries, repos *repository.RepositoryCollection) {
	listings := mustAttachTemplateListings(ctx, ts, repos)
	for _, tl := range listings {
		mustFillTemplateListing(ctx, tl, repos)
	}
}

// *
// * LISTING LISTING LISTING
// *
func mustAttachTemplateListings(ctx context.Context, ts *frontend.TemplateSeries, repos *repository.RepositoryCollection) []*frontend.TemplateListing {
	templateListings := mustBuildTemplateListingsShallow(ctx, ts.Series.Id, repos)
	ts.TemplateListings = make([]frontend.Templatable, len(templateListings))
	for i, templateListing := range templateListings {
		ts.TemplateListings[i] = templateListing
	}
	return templateListings
}

func mustBuildTemplateListingsShallow(ctx context.Context, parentSeriesId int64, repos *repository.RepositoryCollection) []*frontend.TemplateListing {
	listings, err := repos.ListingRepo.GetListingsByParentSeriesIdOrderByPositionAscending(ctx, parentSeriesId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Listings for Series")
	}

	return mustBuildTemplateListings(ctx, listings, repos)
}

func mustBuildTemplateListings(ctx context.Context, listings []repository.Listing, repos *repository.RepositoryCollection) []*frontend.TemplateListing {
	templateListings := make([]*frontend.TemplateListing, 0, len(listings))
	for _, listing := range listings {

		var templateWorkOrSeries frontend.Listable
		switch listing.ListingType {
		case "work":
			work, err := repos.WorkRepo.GetOneWorkByListingId(ctx, listing.Id)
			if err != nil {
				log.Println(err)
				panic("unexpected error getting Work for Listing")
			}

			templateWorkOrSeries = mustBuildTemplateWorkShallow(ctx, work.Id, repos)
		case "series":
			series, err := repos.SeriesRepo.GetOneSeriesByListingId(ctx, listing.Id)
			if err != nil {
				log.Println(err)
				panic("unexpected error getting Series for Listing")
			}

			templateWorkOrSeries = mustBuildTemplateSeriesShallow(ctx, series.Id, repos)
		default:
			panic("unexpected Listing type")
		}

		templateListings = append(templateListings, &frontend.TemplateListing{
			Listing:              &listing,
			TemplateWorkOrSeries: templateWorkOrSeries,
		})
	}

	return templateListings
}
func mustFillTemplateListing(ctx context.Context, tl *frontend.TemplateListing, repos *repository.RepositoryCollection) {
	switch v := tl.TemplateWorkOrSeries.(type) {
	case *frontend.TemplateWork:
		mustFillTemplateWork(ctx, v, repos)
	case *frontend.TemplateSeries:
		mustFillTemplateSeries(ctx, v, repos) // recurse
	}
}

// *
// * WORK WORK WORK
// *
func mustInsertNewTemplateWorkListingInInstance(ctx context.Context, tx pgx.Tx, instanceId int64, repos *repository.RepositoryCollection) *frontend.TemplateListing {
	listing, err := repos.ListingRepo.AppendListingToInstanceTx(ctx, tx, instanceId, "work")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Listing")
	}

	templateWork := mustInsertNewTemplateWork(ctx, tx, listing.Id, repos)
	return &frontend.TemplateListing{
		Listing:              &listing,
		TemplateWorkOrSeries: templateWork,
	}
}

func mustInsertNewTemplateWorkListing(ctx context.Context, tx pgx.Tx, parentSeriesId int64, repos *repository.RepositoryCollection) *frontend.TemplateListing {
	listing, err := repos.ListingRepo.AppendListingToSeriesTx(ctx, tx, parentSeriesId, "work")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Listing")
	}
	templateWork := mustInsertNewTemplateWork(ctx, tx, listing.Id, repos)
	return &frontend.TemplateListing{
		Listing:              &listing,
		TemplateWorkOrSeries: templateWork,
	}
}

func mustInsertNewTemplateWork(ctx context.Context, tx pgx.Tx, listingId int64, repos *repository.RepositoryCollection) *frontend.TemplateWork {
	work, err := repos.WorkRepo.InsertWorkTx(ctx, tx, &repository.WorkArguments{
		ListingId: listingId,
		Title:     "Untitled Work",
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Work")
	}

	return &frontend.TemplateWork{
		Work:             &work,
		TemplateContents: nil,
	}
}

func mustBuildTemplateWorkShallow(ctx context.Context, workId int64, repos *repository.RepositoryCollection) *frontend.TemplateWork {
	work, err := repos.WorkRepo.GetOneWorkById(ctx, workId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Work")
	}

	return &frontend.TemplateWork{
		Work:             &work,
		TemplateContents: nil,
	}
}

func mustFillTemplateWork(ctx context.Context, tw *frontend.TemplateWork, repos *repository.RepositoryCollection) {
	contents, err := repos.ContentRepo.GetContentsByWorkIdOrderByPositionAscending(ctx, tw.Work.Id)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Contents for Work")
	}

	templateContents := mustBuildTemplateContents(ctx, contents, repos)
	for _, templateContent := range templateContents {
		mustFillTemplateContent(ctx, templateContent, repos)
	}
	tw.TemplateContents = make([]frontend.Templatable, len(templateContents))
	for i, templateContent := range templateContents {
		tw.TemplateContents[i] = templateContent
	}
}

// *
// * GROUP GROUP GROUP
// *
func mustInsertNewTemplateGroupContentInWork(ctx context.Context, tx pgx.Tx, workId int64, repos *repository.RepositoryCollection) *frontend.TemplateContent {
	content, err := repos.ContentRepo.AppendContentToWorkTx(ctx, tx, workId, "group")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Content")
	}

	templateGroup := mustInsertNewNestedTemplateGroup(ctx, tx, content.Id, repos)
	return &frontend.TemplateContent{
		Content:                    &content,
		TemplateGroupOrTextOrMedia: templateGroup,
	}
}

func mustInsertNewTemplateGroupContent(ctx context.Context, tx pgx.Tx, parentGroupId int64, repos *repository.RepositoryCollection) *frontend.TemplateContent {
	content, err := repos.ContentRepo.AppendContentToGroupTx(ctx, tx, parentGroupId, "group")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Content")
	}

	templateGroup := mustInsertNewNestedTemplateGroup(ctx, tx, content.Id, repos)

	return &frontend.TemplateContent{
		Content:                    &content,
		TemplateGroupOrTextOrMedia: templateGroup,
	}
}

func mustInsertNewNestedTemplateGroup(ctx context.Context, tx pgx.Tx, contentId int64, repos *repository.RepositoryCollection) *frontend.TemplateGroup {
	group, err := repos.GroupRepo.InsertGroupTx(ctx, tx, &repository.GroupArguments{
		ContentId:     contentId,
		SwapDirection: false,
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Group")
	}

	return &frontend.TemplateGroup{
		Group:            &group,
		TemplateContents: nil,
	}
}

func mustBuildTemplateGroupShallow(ctx context.Context, groupId int64, repos *repository.RepositoryCollection) *frontend.TemplateGroup {
	group, err := repos.GroupRepo.GetOneGroupById(ctx, groupId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Group")
	}

	return &frontend.TemplateGroup{
		Group:            &group,
		TemplateContents: nil,
	}
}

func mustFillTemplateGroup(ctx context.Context, tg *frontend.TemplateGroup, repos *repository.RepositoryCollection) {
	contents := mustAttachTemplateContents(ctx, tg, repos)
	for _, tc := range contents {
		mustFillTemplateContent(ctx, tc, repos)
	}
}

// *
// * CONTENT CONTENT CONTENT
// *
func mustAttachTemplateContents(ctx context.Context, tg *frontend.TemplateGroup, repos *repository.RepositoryCollection) []*frontend.TemplateContent {
	templateContents := mustBuildTemplateContentsShallow(ctx, tg.Group.Id, repos)
	tg.TemplateContents = make([]frontend.Templatable, len(templateContents))
	for i, templateContent := range templateContents {
		tg.TemplateContents[i] = templateContent
	}
	return templateContents
}

func mustBuildTemplateContentsShallow(ctx context.Context, parentGroupId int64, repos *repository.RepositoryCollection) []*frontend.TemplateContent {
	contents, err := repos.ContentRepo.GetContentsByParentGroupIdOrderByPositionAscending(ctx, parentGroupId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Contents for Group")
	}

	return mustBuildTemplateContents(ctx, contents, repos)
}

func mustBuildTemplateContents(ctx context.Context, contents []repository.Content, repos *repository.RepositoryCollection) []*frontend.TemplateContent {
	templateContents := make([]*frontend.TemplateContent, 0, len(contents))
	for _, content := range contents {

		var templateGroupOrTextOrMedia frontend.Templatable
		switch content.ContentType {
		case "group":
			group, err := repos.GroupRepo.GetOneGroupByContentId(ctx, content.Id)
			if err != nil {
				log.Println(err)
				panic("unexpected error getting Group for Content")
			}

			templateGroupOrTextOrMedia = mustBuildTemplateGroupShallow(ctx, group.Id, repos)
		case "text":
			text, err := repos.TextRepo.GetOneTextByContentId(ctx, content.Id)
			if err != nil {
				log.Println(err)
				panic("unexpected error getting Text for Content")
			}

			templateGroupOrTextOrMedia = mustBuildTemplateTextShallow(ctx, text.Id, repos)
		case "media":
			media, err := repos.MediaRepo.GetOneMediaByContentId(ctx, content.Id)
			if err != nil {
				log.Println(err)
				panic("unexpected error getting Media for Content")
			}

			templateGroupOrTextOrMedia = mustBuildTemplateMediaShallow(ctx, media.Id, repos)
		default:
			panic("unexpected Content type")
		}

		templateContents = append(templateContents, &frontend.TemplateContent{
			Content:                    &content,
			TemplateGroupOrTextOrMedia: templateGroupOrTextOrMedia,
		})
	}

	return templateContents
}
func mustFillTemplateContent(ctx context.Context, tc *frontend.TemplateContent, repos *repository.RepositoryCollection) {
	switch v := tc.TemplateGroupOrTextOrMedia.(type) {
	case *frontend.TemplateGroup:
		mustFillTemplateGroup(ctx, v, repos) // recurse
	case *frontend.TemplateText:
		// leaf, nothing to do
	case *frontend.TemplateMedia:
		// leaf, storage data was attached while building it
	}
}

// *
// * TEXT TEXT TEXT
// *
func mustInsertNewTemplateTextContentInWork(ctx context.Context, tx pgx.Tx, workId int64, repos *repository.RepositoryCollection) *frontend.TemplateContent {
	content, err := repos.ContentRepo.AppendContentToWorkTx(ctx, tx, workId, "text")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Content")
	}

	templateText := mustInsertNewTemplateText(ctx, tx, content.Id, repos)
	return &frontend.TemplateContent{
		Content:                    &content,
		TemplateGroupOrTextOrMedia: templateText,
	}
}

func mustInsertNewTemplateTextContent(ctx context.Context, tx pgx.Tx, parentGroupId int64, repos *repository.RepositoryCollection) *frontend.TemplateContent {
	content, err := repos.ContentRepo.AppendContentToGroupTx(ctx, tx, parentGroupId, "text")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Content")
	}

	templateText := mustInsertNewTemplateText(ctx, tx, content.Id, repos)

	return &frontend.TemplateContent{
		Content:                    &content,
		TemplateGroupOrTextOrMedia: templateText,
	}
}

func mustInsertNewTemplateText(ctx context.Context, tx pgx.Tx, contentId int64, repos *repository.RepositoryCollection) *frontend.TemplateText {
	text, err := repos.TextRepo.InsertTextTx(ctx, tx, &repository.TextArguments{
		ContentId: contentId,
		Content:   "",
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Group")
	}

	return &frontend.TemplateText{
		Text: &text,
	}
}

// Text is a leaf — "shallow" is also "full", but named for consistency
// with how it's referenced from mustBuildTemplateContentsShallow.
func mustBuildTemplateTextShallow(ctx context.Context, textId int64, repos *repository.RepositoryCollection) *frontend.TemplateText {
	text, err := repos.TextRepo.GetOneTextById(ctx, textId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Text")
	}

	return &frontend.TemplateText{
		Text: &text,
	}
}

// *
// * MEDIA MEDIA MEDIA
// *
func mustInsertNewTemplateMediaContentInWork(ctx context.Context, tx pgx.Tx, workId int64, repos *repository.RepositoryCollection) *frontend.TemplateContent {
	content, err := repos.ContentRepo.AppendContentToWorkTx(ctx, tx, workId, "media")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Content")
	}

	templateMedia := mustInsertNewTemplateMedia(ctx, tx, content.Id, repos)
	return &frontend.TemplateContent{
		Content:                    &content,
		TemplateGroupOrTextOrMedia: templateMedia,
	}
}

func mustInsertNewTemplateMediaContent(ctx context.Context, tx pgx.Tx, parentGroupId int64, repos *repository.RepositoryCollection) *frontend.TemplateContent {
	content, err := repos.ContentRepo.AppendContentToGroupTx(ctx, tx, parentGroupId, "media")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Content")
	}

	templateMedia := mustInsertNewTemplateMedia(ctx, tx, content.Id, repos)

	return &frontend.TemplateContent{
		Content:                    &content,
		TemplateGroupOrTextOrMedia: templateMedia,
	}
}

func mustInsertNewTemplateMedia(ctx context.Context, tx pgx.Tx, contentId int64, repos *repository.RepositoryCollection) *frontend.TemplateMedia {
	media, err := repos.MediaRepo.InsertMediaTx(ctx, tx, &repository.MediaArguments{
		ContentId: contentId,
		FileHash:  nil,
		Caption:   nil,
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Media")
	}

	return &frontend.TemplateMedia{
		Media: &media,
	}
}

func mustBuildTemplateMediaShallow(ctx context.Context, mediaId int64, repos *repository.RepositoryCollection) *frontend.TemplateMedia {
	media, err := repos.MediaRepo.GetOneMediaById(ctx, mediaId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Media")
	}

	templateMedia := &frontend.TemplateMedia{Media: &media}
	if media.FileHash == nil {
		return templateMedia
	}

	file, err := repos.FileRepo.GetOneFileByHash(ctx, *media.FileHash)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting File for Media")
	}

	filenode, err := repos.FilenodeRepo.GetOneFilenodeByFileIdOrderByIdAscending(ctx, file.Id)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Filenode for File")
	}

	fileserver, err := repos.FileserverRepo.GetOneFileserverById(ctx, filenode.FileserverId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Fileserver for Filenode")
	}

	templateMedia.File = &file
	templateMedia.Filenode = &filenode
	templateMedia.Fileserver = &fileserver
	return templateMedia
}
