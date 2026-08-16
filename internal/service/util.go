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

	templateCollection := mustInsertNewTemplateCollection(ctx, tx, instance.Id, repos)

	return &frontend.TemplateInstance{
		Instance:           &instance,
		TemplateCollection: templateCollection, // keep the subtree we already built
	}
}

// The Instance is resolved once at startup, so the caller already holds the row
// and nothing has to be fetched here.
func buildTemplateInstanceShallow(instance *repository.Instance) *frontend.TemplateInstance {
	return &frontend.TemplateInstance{
		Instance:           instance,
		TemplateCollection: nil,
	}
}

// *
// * COLLECTION COLLECTION COLLECTION
// *
func mustInsertNewTemplateCollection(ctx context.Context, tx pgx.Tx, instanceId int64, repos *repository.RepositoryCollection) *frontend.TemplateCollection {
	collection, err := repos.CollectionRepo.InsertCollectionTx(ctx, tx, &repository.CollectionArguments{
		InstanceId: instanceId,
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Collection")
	}

	templateSeries := mustInsertNewRootTemplateSeries(ctx, tx, collection.Id, repos)

	return &frontend.TemplateCollection{
		Collection:     &collection,
		TemplateSeries: templateSeries, // keep what we built, don't discard it
	}
}

func mustAttachTemplateCollection(ctx context.Context, ti *frontend.TemplateInstance, repos *repository.RepositoryCollection) *frontend.TemplateCollection {
	collection, err := repos.CollectionRepo.GetOneCollectionByInstanceId(ctx, ti.Instance.Id)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Collection")
	}

	templateCollection := &frontend.TemplateCollection{
		Collection:     &collection,
		TemplateSeries: nil,
	}

	ti.TemplateCollection = templateCollection
	return templateCollection
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

// A Series's parent is either a Collection (root) or a Listing (nested), never
// both and never neither — mirroring the CHECK constraint on the series table.
func mustInsertNewRootTemplateSeries(ctx context.Context, tx pgx.Tx, collectionId int64, repos *repository.RepositoryCollection) *frontend.TemplateSeries {
	series, err := repos.SeriesRepo.InsertSeriesTx(ctx, tx, &repository.SeriesArguments{
		CollectionId: &collectionId,
		ListingId:    nil,
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

func mustInsertNewNestedTemplateSeries(ctx context.Context, tx pgx.Tx, listingId int64, repos *repository.RepositoryCollection) *frontend.TemplateSeries {
	series, err := repos.SeriesRepo.InsertSeriesTx(ctx, tx, &repository.SeriesArguments{
		CollectionId: nil,
		ListingId:    &listingId,
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

// Only root Series carry a collection_id, so this attaches the Collection's root Series.
func mustAttachTemplateSeries(ctx context.Context, tc *frontend.TemplateCollection, repos *repository.RepositoryCollection) *frontend.TemplateSeries {
	series, err := repos.SeriesRepo.GetOneSeriesByCollectionId(ctx, tc.Collection.Id)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Series")
	}

	templateSeries := &frontend.TemplateSeries{
		Series:           &series,
		TemplateListings: nil,
	}

	tc.TemplateSeries = templateSeries
	return templateSeries
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

	templatables := make([]frontend.Templatable, len(templateListings))
	for i, tl := range templateListings {
		templatables[i] = tl
	}

	ts.TemplateListings = templatables
	return templateListings
}

func mustBuildTemplateListingsShallow(ctx context.Context, parentSeriesId int64, repos *repository.RepositoryCollection) []*frontend.TemplateListing {
	listings, err := repos.ListingRepo.GetListingsByParentSeriesIdOrderByPositionAscending(ctx, parentSeriesId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Listings for Series")
	}

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

	templateCanvas := mustInsertNewTemplateCanvas(ctx, tx, work.Id, repos)

	return &frontend.TemplateWork{
		Work:           &work,
		TemplateCanvas: templateCanvas,
	}
}

func mustBuildTemplateWorkShallow(ctx context.Context, workId int64, repos *repository.RepositoryCollection) *frontend.TemplateWork {
	work, err := repos.WorkRepo.GetOneWorkById(ctx, workId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Work")
	}

	return &frontend.TemplateWork{
		Work:           &work,
		TemplateCanvas: nil,
	}
}

func mustFillTemplateWork(ctx context.Context, tw *frontend.TemplateWork, repos *repository.RepositoryCollection) {
	tc := mustAttachTemplateCanvas(ctx, tw, repos)
	mustFillTemplateCanvas(ctx, tc, repos)
}

// *
// * CANVAS CANVAS CANVAS
// *
func mustInsertNewTemplateCanvas(ctx context.Context, tx pgx.Tx, workId int64, repos *repository.RepositoryCollection) *frontend.TemplateCanvas {
	canvas, err := repos.CanvasRepo.InsertCanvasTx(ctx, tx, &repository.CanvasArguments{
		WorkId: workId,
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Canvas")
	}

	templateGroup := mustInsertNewRootTemplateGroup(ctx, tx, canvas.Id, repos)

	return &frontend.TemplateCanvas{
		Canvas:        &canvas,
		TemplateGroup: templateGroup, // keep the subtree we already built
	}
}

func mustAttachTemplateCanvas(ctx context.Context, tw *frontend.TemplateWork, repos *repository.RepositoryCollection) *frontend.TemplateCanvas {
	canvas, err := repos.CanvasRepo.GetOneCanvasByWorkId(ctx, tw.Work.Id)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Canvas")
	}

	templateCanvas := &frontend.TemplateCanvas{
		Canvas:        &canvas,
		TemplateGroup: nil,
	}

	tw.TemplateCanvas = templateCanvas
	return templateCanvas
}

func mustFillTemplateCanvas(ctx context.Context, tc *frontend.TemplateCanvas, repos *repository.RepositoryCollection) {
	tg := mustAttachTemplateGroup(ctx, tc, repos)
	mustFillTemplateGroup(ctx, tg, repos)
}

// *
// * GROUP GROUP GROUP
// *
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

// A Group's parent is either a Canvas (root) or a Content (nested), never both
// and never neither — mirroring the CHECK constraint on the groups table.
func mustInsertNewRootTemplateGroup(ctx context.Context, tx pgx.Tx, canvasId int64, repos *repository.RepositoryCollection) *frontend.TemplateGroup {
	group, err := repos.GroupRepo.InsertGroupTx(ctx, tx, &repository.GroupArguments{
		CanvasId:      &canvasId,
		ContentId:     nil,
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

func mustInsertNewNestedTemplateGroup(ctx context.Context, tx pgx.Tx, contentId int64, repos *repository.RepositoryCollection) *frontend.TemplateGroup {
	group, err := repos.GroupRepo.InsertGroupTx(ctx, tx, &repository.GroupArguments{
		CanvasId:      nil,
		ContentId:     &contentId,
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

func mustAttachTemplateGroup(ctx context.Context, tc *frontend.TemplateCanvas, repos *repository.RepositoryCollection) *frontend.TemplateGroup {
	group, err := repos.GroupRepo.GetOneGroupByCanvasId(ctx, tc.Canvas.Id)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Group")
	}

	templateGroup := &frontend.TemplateGroup{
		Group:            &group,
		TemplateContents: nil,
	}

	tc.TemplateGroup = templateGroup
	return templateGroup
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

	templatables := make([]frontend.Templatable, len(templateContents))
	for i, tc := range templateContents {
		templatables[i] = tc
	}

	tg.TemplateContents = templatables
	return templateContents
}

func mustBuildTemplateContentsShallow(ctx context.Context, parentGroupId int64, repos *repository.RepositoryCollection) []*frontend.TemplateContent {
	contents, err := repos.ContentRepo.GetContentsByParentGroupIdOrderByPositionAscending(ctx, parentGroupId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Contents for Group")
	}

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
		mustAttachTemplateCaption(ctx, v, repos)
	}
}

// *
// * TEXT TEXT TEXT
// *
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
// Sources are always fetched (cheap, always wanted); TemplateCaption is left
// nil since a Media may have no Caption — attach separately.
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

	sources, err := repos.SourceRepo.GetSourcesByMediaId(ctx, media.Id)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Sources for Media")
	}

	templateSources := make([]frontend.Templatable, 0, len(sources))
	for i := range sources {
		templateSources = append(templateSources, mustBuildTemplateSource(ctx, &sources[i], repos))
	}

	return &frontend.TemplateMedia{
		Media:           &media,
		TemplateSources: templateSources,
		TemplateCaption: nil,
	}
}

// *
// * CAPTION CAPTION CAPTION
// *
// A Caption belongs to a single Media — captions.media_id is unique — so it is
// inserted against the Media it captions rather than appended to a Group.
func mustInsertNewTemplateCaption(ctx context.Context, tx pgx.Tx, mediaId int64, repos *repository.RepositoryCollection) *frontend.TemplateCaption {
	caption, err := repos.CaptionRepo.InsertCaptionTx(ctx, tx, &repository.CaptionArguments{
		MediaId: mediaId,
		Content: "",
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Caption")
	}

	return &frontend.TemplateCaption{
		Caption: &caption,
	}
}

// Nil if the Media has no caption. Leaves tm.TemplateCaption untouched in that
// case: assigning a nil *TemplateCaption would make the Templatable interface
// non-nil, and templates guarding on it would then dereference nil.
func mustAttachTemplateCaption(ctx context.Context, tm *frontend.TemplateMedia, repos *repository.RepositoryCollection) *frontend.TemplateCaption {
	caption, err := repos.CaptionRepo.GetOptionalCaptionByMediaId(ctx, tm.Media.Id)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Caption")
	}

	if caption == nil {
		return nil
	}

	templateCaption := &frontend.TemplateCaption{
		Caption: caption,
	}

	tm.TemplateCaption = templateCaption
	return templateCaption
}

// *
// * SOURCE SOURCE SOURCE
// *
// A Source names a stored file rather than owning one: it points at a Filename,
// which belongs to a File, which is stored on one or more Fileservers as
// Filenodes. Rendering a Source needs the whole walk, because the url is
// assembled from the Fileserver and the path the Filenode has on it.
func mustBuildTemplateSource(ctx context.Context, source *repository.Source, repos *repository.RepositoryCollection) *frontend.TemplateSource {
	filename, err := repos.FilenameRepo.GetOneFilenameById(ctx, source.FilenameId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Filename for Source")
	}

	file, err := repos.FileRepo.GetOneFileById(ctx, filename.FileId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting File for Filename")
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

	return &frontend.TemplateSource{
		Source:     source,
		Filename:   &filename,
		File:       &file,
		Filenode:   &filenode,
		Fileserver: &fileserver,
	}
}
