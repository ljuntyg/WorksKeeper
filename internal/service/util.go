package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/template/frontend"
	"log"

	"github.com/jackc/pgx/v5"
)

// *
// * COLLECTION COLLECTION COLLECTION
// *
func mustInsertNewTemplateCollection(tx pgx.Tx, repos *repository.RepositoryCollection) *frontend.TemplateCollection {
	rootSeries := mustInsertNewTemplateSeries(tx, nil, repos)

	collection, err := repos.CollectionRepo.InsertCollectionTx(tx, &repository.CollectionArguments{
		RootSeriesId: rootSeries.Series.Id,
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Collection")
	}

	return &frontend.TemplateCollection{
		Collection:     &collection,
		TemplateSeries: rootSeries, // keep what we built, don't discard it
	}
}

func mustBuildTemplateCollectionShallow(collectionId int64, repos *repository.RepositoryCollection) *frontend.TemplateCollection {
	collection, err := repos.CollectionRepo.GetCollection(collectionId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Collection")
	}

	return &frontend.TemplateCollection{
		Collection:     &collection,
		TemplateSeries: nil,
	}
}

// *
// * SERIES SERIES SERIES
// *
func mustInsertNewTemplateSeriesListing(tx pgx.Tx, parentSeriesId int64, repos *repository.RepositoryCollection) *frontend.TemplateListing {
	listing, err := repos.ListingRepo.InsertListingAppendTx(tx, parentSeriesId, "series")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Listing")
	}

	listingId := listing.Id
	templateSeries := mustInsertNewTemplateSeries(tx, &listingId, repos)

	return &frontend.TemplateListing{
		Listing:              &listing,
		TemplateWorkOrSeries: templateSeries,
	}
}

func mustInsertNewTemplateSeries(tx pgx.Tx, listingId *int64, repos *repository.RepositoryCollection) *frontend.TemplateSeries {
	series, err := repos.SeriesRepo.InsertSeriesTx(tx, &repository.SeriesArguments{
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

func mustAttachTemplateSeries(tc *frontend.TemplateCollection, repos *repository.RepositoryCollection) *frontend.TemplateSeries {
	templateSeries := mustBuildTemplateSeriesShallow(tc.Collection.RootSeriesId, repos)
	tc.TemplateSeries = templateSeries
	return templateSeries
}

func mustBuildTemplateSeriesShallow(seriesId int64, repos *repository.RepositoryCollection) *frontend.TemplateSeries {
	series, err := repos.SeriesRepo.GetSeries(seriesId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Series")
	}

	return &frontend.TemplateSeries{
		Series:           &series,
		TemplateListings: nil,
	}
}

func mustFillTemplateSeries(ts *frontend.TemplateSeries, repos *repository.RepositoryCollection) {
	listings := mustAttachTemplateListings(ts, repos)
	for _, tl := range listings {
		mustFillTemplateListing(tl, repos)
	}
}

// *
// * LISTING LISTING LISTING
// *
func mustAttachTemplateListings(ts *frontend.TemplateSeries, repos *repository.RepositoryCollection) []*frontend.TemplateListing {
	templateListings := mustBuildTemplateListingsShallow(ts.Series.Id, repos)

	templatables := make([]frontend.Templatable, len(templateListings))
	for i, tl := range templateListings {
		templatables[i] = tl
	}

	ts.TemplateListings = templatables
	return templateListings
}

func mustBuildTemplateListingsShallow(parentSeriesId int64, repos *repository.RepositoryCollection) []*frontend.TemplateListing {
	listings, err := repos.ListingRepo.GetListingsByParentSeriesIdOrderByPositionAscending(parentSeriesId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Listings for Series")
	}

	templateListings := make([]*frontend.TemplateListing, 0, len(listings))
	for _, listing := range listings {

		var templateWorkOrSeries frontend.Listable
		switch listing.ListingType {
		case "work":
			work, err := repos.WorkRepo.GetWorkByListingId(listing.Id)
			if err != nil {
				log.Println(err)
				panic("unexpected error getting Work for Listing")
			}

			templateWorkOrSeries = mustBuildTemplateWorkShallow(work.Id, repos)
		case "series":
			series, err := repos.SeriesRepo.GetSeriesByListingId(listing.Id)
			if err != nil {
				log.Println(err)
				panic("unexpected error getting Series for Listing")
			}

			templateWorkOrSeries = mustBuildTemplateSeriesShallow(series.Id, repos)
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

func mustFillTemplateListing(tl *frontend.TemplateListing, repos *repository.RepositoryCollection) {
	switch v := tl.TemplateWorkOrSeries.(type) {
	case *frontend.TemplateWork:
		mustFillTemplateWork(v, repos)
	case *frontend.TemplateSeries:
		mustFillTemplateSeries(v, repos) // recurse
	}
}

// *
// * WORK WORK WORK
// *
func mustInsertNewTemplateWorkListing(tx pgx.Tx, parentSeriesId int64, repos *repository.RepositoryCollection) *frontend.TemplateListing {
	listing, err := repos.ListingRepo.InsertListingAppendTx(tx, parentSeriesId, "work")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Listing")
	}
	templateWork := mustInsertNewTemplateWork(tx, listing.Id, repos)
	return &frontend.TemplateListing{
		Listing:              &listing,
		TemplateWorkOrSeries: templateWork,
	}
}

func mustInsertNewTemplateWork(tx pgx.Tx, listingId int64, repos *repository.RepositoryCollection) *frontend.TemplateWork {
	work, err := repos.WorkRepo.InsertWorkTx(tx, &repository.WorkArguments{
		ListingId: listingId,
		Title:     "Untitled Work",
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Work")
	}

	templateCanvas := mustInsertNewTemplateCanvas(tx, work.Id, repos)

	return &frontend.TemplateWork{
		Work:           &work,
		TemplateCanvas: templateCanvas,
	}
}

func mustBuildTemplateWorkShallow(workId int64, repos *repository.RepositoryCollection) *frontend.TemplateWork {
	work, err := repos.WorkRepo.GetWork(workId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Work")
	}

	return &frontend.TemplateWork{
		Work:           &work,
		TemplateCanvas: nil,
	}
}

func mustFillTemplateWork(tw *frontend.TemplateWork, repos *repository.RepositoryCollection) {
	tc := mustAttachTemplateCanvas(tw, repos)
	mustFillTemplateCanvas(tc, repos)
}

// *
// * CANVAS CANVAS CANVAS
// *
func mustInsertNewTemplateCanvas(tx pgx.Tx, workId int64, repos *repository.RepositoryCollection) *frontend.TemplateCanvas {
	canvas, err := repos.CanvasRepo.InsertCanvasTx(tx, &repository.CanvasArguments{
		WorkId: workId,
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Canvas")
	}

	templateGroup := mustInsertNewRootTemplateGroup(tx, canvas.Id, repos)

	return &frontend.TemplateCanvas{
		Canvas:        &canvas,
		TemplateGroup: templateGroup, // keep the subtree we already built
	}
}

func mustAttachTemplateCanvas(tw *frontend.TemplateWork, repos *repository.RepositoryCollection) *frontend.TemplateCanvas {
	canvas, err := repos.CanvasRepo.GetCanvasByWorkId(tw.Work.Id)
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

func mustFillTemplateCanvas(tc *frontend.TemplateCanvas, repos *repository.RepositoryCollection) {
	tg := mustAttachTemplateGroup(tc, repos)
	mustFillTemplateGroup(tg, repos)
}

// *
// * GROUP GROUP GROUP
// *
func mustInsertNewTemplateGroupContent(tx pgx.Tx, parentGroupId int64, repos *repository.RepositoryCollection) *frontend.TemplateContent {
	content, err := repos.ContentRepo.InsertContentAppendTx(tx, parentGroupId, "group")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Content")
	}

	templateGroup := mustInsertNewNestedTemplateGroup(tx, content.Id, repos)

	return &frontend.TemplateContent{
		Content:                    &content,
		TemplateGroupOrTextOrMedia: templateGroup,
	}
}

// A Group's parent is either a Canvas (root) or a Content (nested), never both
// and never neither — mirroring the CHECK constraint on the groups table.
func mustInsertNewRootTemplateGroup(tx pgx.Tx, canvasId int64, repos *repository.RepositoryCollection) *frontend.TemplateGroup {
	group, err := repos.GroupRepo.InsertGroupTx(tx, &repository.GroupArguments{
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

func mustInsertNewNestedTemplateGroup(tx pgx.Tx, contentId int64, repos *repository.RepositoryCollection) *frontend.TemplateGroup {
	group, err := repos.GroupRepo.InsertGroupTx(tx, &repository.GroupArguments{
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

func mustAttachTemplateGroup(tc *frontend.TemplateCanvas, repos *repository.RepositoryCollection) *frontend.TemplateGroup {
	group, err := repos.GroupRepo.GetRootGroupByCanvasId(tc.Canvas.Id)
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

func mustBuildTemplateGroupShallow(groupId int64, repos *repository.RepositoryCollection) *frontend.TemplateGroup {
	group, err := repos.GroupRepo.GetGroup(groupId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Group")
	}

	return &frontend.TemplateGroup{
		Group:            &group,
		TemplateContents: nil,
	}
}

func mustFillTemplateGroup(tg *frontend.TemplateGroup, repos *repository.RepositoryCollection) {
	contents := mustAttachTemplateContents(tg, repos)
	for _, tc := range contents {
		mustFillTemplateContent(tc, repos)
	}
}

// *
// * CONTENT CONTENT CONTENT
// *
func mustAttachTemplateContents(tg *frontend.TemplateGroup, repos *repository.RepositoryCollection) []*frontend.TemplateContent {
	templateContents := mustBuildTemplateContentsShallow(tg.Group.Id, repos)

	templatables := make([]frontend.Templatable, len(templateContents))
	for i, tc := range templateContents {
		templatables[i] = tc
	}

	tg.TemplateContents = templatables
	return templateContents
}

func mustBuildTemplateContentsShallow(parentGroupId int64, repos *repository.RepositoryCollection) []*frontend.TemplateContent {
	contents, err := repos.ContentRepo.GetContentsByParentGroupIdOrderByPositionAscending(parentGroupId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Contents for Group")
	}

	templateContents := make([]*frontend.TemplateContent, 0, len(contents))
	for _, content := range contents {

		var templateGroupOrTextOrMedia frontend.Templatable
		switch content.ContentType {
		case "group":
			group, err := repos.GroupRepo.GetGroupByContentId(content.Id)
			if err != nil {
				log.Println(err)
				panic("unexpected error getting Group for Content")
			}

			templateGroupOrTextOrMedia = mustBuildTemplateGroupShallow(group.Id, repos)
		case "text":
			text, err := repos.TextRepo.GetTextByContentId(content.Id)
			if err != nil {
				log.Println(err)
				panic("unexpected error getting Text for Content")
			}

			templateGroupOrTextOrMedia = mustBuildTemplateTextShallow(text.Id, repos)
		case "media":
			media, err := repos.MediaRepo.GetMediaByContentId(content.Id)
			if err != nil {
				log.Println(err)
				panic("unexpected error getting Media for Content")
			}

			templateGroupOrTextOrMedia = mustBuildTemplateMediaShallow(media.Id, repos)
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

func mustFillTemplateContent(tc *frontend.TemplateContent, repos *repository.RepositoryCollection) {
	switch v := tc.TemplateGroupOrTextOrMedia.(type) {
	case *frontend.TemplateGroup:
		mustFillTemplateGroup(v, repos) // recurse
	case *frontend.TemplateText:
		// leaf, nothing to do
	case *frontend.TemplateMedia:
		mustAttachTemplateCaption(v, repos)
	}
}

// *
// * TEXT TEXT TEXT
// *
func mustInsertNewTemplateTextContent(tx pgx.Tx, parentGroupId int64, repos *repository.RepositoryCollection) *frontend.TemplateContent {
	content, err := repos.ContentRepo.InsertContentAppendTx(tx, parentGroupId, "text")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Content")
	}

	templateText := mustInsertNewTemplateText(tx, content.Id, repos)

	return &frontend.TemplateContent{
		Content:                    &content,
		TemplateGroupOrTextOrMedia: templateText,
	}
}

func mustInsertNewTemplateText(tx pgx.Tx, contentId int64, repos *repository.RepositoryCollection) *frontend.TemplateText {
	text, err := repos.TextRepo.InsertTextTx(tx, &repository.TextArguments{
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
func mustBuildTemplateTextShallow(textId int64, repos *repository.RepositoryCollection) *frontend.TemplateText {
	text, err := repos.TextRepo.GetText(textId)
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
func mustInsertNewTemplateMediaContent(tx pgx.Tx, parentGroupId int64, repos *repository.RepositoryCollection) *frontend.TemplateContent {
	content, err := repos.ContentRepo.InsertContentAppendTx(tx, parentGroupId, "media")
	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Content")
	}

	templateMedia := mustInsertNewTemplateMedia(tx, content.Id, repos)

	return &frontend.TemplateContent{
		Content:                    &content,
		TemplateGroupOrTextOrMedia: templateMedia,
	}
}

func mustInsertNewTemplateMedia(tx pgx.Tx, contentId int64, repos *repository.RepositoryCollection) *frontend.TemplateMedia {
	media, err := repos.MediaRepo.InsertMediaTx(tx, &repository.MediaArguments{
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

func mustBuildTemplateMediaShallow(mediaId int64, repos *repository.RepositoryCollection) *frontend.TemplateMedia {
	media, err := repos.MediaRepo.GetMedia(mediaId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Media")
	}

	sources, err := repos.SourceRepo.GetSourcesByMediaId(media.Id)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Sources for Media")
	}

	templateSources := make([]frontend.Templatable, 0, len(sources))
	for i := range sources {
		templateSources = append(templateSources, mustBuildTemplateSource(&sources[i], repos))
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
func mustInsertNewTemplateCaption(tx pgx.Tx, mediaId int64, repos *repository.RepositoryCollection) *frontend.TemplateCaption {
	caption, err := repos.CaptionRepo.InsertCaptionTx(tx, &repository.CaptionArguments{
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
func mustAttachTemplateCaption(tm *frontend.TemplateMedia, repos *repository.RepositoryCollection) *frontend.TemplateCaption {
	caption, err := repos.CaptionRepo.GetCaptionByMediaId(tm.Media.Id)
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
func mustBuildTemplateSource(source *repository.Source, repos *repository.RepositoryCollection) *frontend.TemplateSource {
	filename, err := repos.FilenameRepo.GetFilename(source.FilenameId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Filename for Source")
	}

	file, err := repos.FileRepo.GetFile(filename.FileId)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting File for Filename")
	}

	filenode, err := repos.FilenodeRepo.GetFilenodeByFileIdOrderByIdAscending(file.Id)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Filenode for File")
	}

	fileserver, err := repos.FileserverRepo.GetFileserver(filenode.FileserverId)
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
