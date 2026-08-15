package handler

import (
	"WorksKeeper/internal/service"
	"log"
	"net/http"
	"strconv"
)

type CanvasHandler struct {
	canvasService *service.CanvasService
}

// maxUploadSize is the largest body a POST is read from. Without a limit a
// single request can fill the disk, because the parts of a multipart form that
// do not fit in memory are spilled to a temporary file however large they are.
const maxUploadSize = 256 << 20

func (ch *CanvasHandler) Init(canvasService *service.CanvasService) {
	ch.canvasService = canvasService
}

func (ch *CanvasHandler) HandleRequest(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ch.handleGet(rw, r, true)
	case http.MethodPost:
		ch.handlePost(rw, r)
	}
}

func (ch *CanvasHandler) HandleRequestNew(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ch.handleGetNew(rw, r)
	}
}

func (ch *CanvasHandler) handleGet(rw http.ResponseWriter, r *http.Request, editing bool) {
	workId := mustNumberingStringToId(r.PathValue("numbering"))
	ch.canvasService.GetTemplateData(r.Context(), workId, editing).ExecuteTemplate(rw)
}

func (ch *CanvasHandler) handlePost(rw http.ResponseWriter, r *http.Request) {
	/*
		OK view-work
		OK edit-work

		OK add-text
		OK add-media
		OK add-group

		OK delete-content
		OK content-neg-dir
		OK content-pos-dir

		OK delete caption
		OK add caption

		OK upload
	*/

	// Bound the body before anything reads it: parsing the form is what pulls
	// an upload in, so the limit has to be in place before that happens.
	r.Body = http.MaxBytesReader(rw, r.Body, maxUploadSize)

	action, value := extractActionAndValueFromRequest(r)
	log.Println(action, value)

	// Every action submits the whole form, so what was typed is saved before
	// the action runs and the page is rendered again from the database.
	ch.saveEdits(r)

	switch action {
	case "edit-work":
		ch.handleGet(rw, r, true)
	case "view-work":
		ch.handleGet(rw, r, false)
	case "add-text":
		groupId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Group id")
		}

		ch.handleAddText(rw, r, int64(groupId))
	case "add-media":
		groupId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Group id")
		}

		ch.handleAddMedia(rw, r, int64(groupId))
	case "add-group":
		groupId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Group id")
		}

		ch.handleAddGroup(rw, r, int64(groupId))
	case "add-caption":
		mediaId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Media id")
		}

		ch.handleAddCaption(rw, r, int64(mediaId))
	case "content-increase-position":
		contentId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Content id")
		}

		ch.handleContentIncreasePosition(rw, r, int64(contentId))
	case "content-decrease-position":
		contentId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Content id")
		}

		ch.handleContentDecreasePosition(rw, r, int64(contentId))
	case "delete-content":
		contentId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Content id")
		}

		ch.handleDeleteContent(rw, r, int64(contentId))
	case "delete-caption":
		captionId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Caption id")
		}

		ch.handleDeleteCaption(rw, r, int64(captionId))
	case "upload":
		mediaId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Media id")
		}

		ch.handleUpload(rw, r, int64(mediaId))
	}
}

func (ch *CanvasHandler) handleGetNew(rw http.ResponseWriter, r *http.Request) {
	work := ch.canvasService.MustInsertNewWorkInBaseCollection(r.Context())
	http.Redirect(rw, r, "/compose"+work.GetNumberingUrlString(), http.StatusFound)
}

func (ch *CanvasHandler) saveEdits(r *http.Request) {
	workId := mustNumberingStringToId(r.PathValue("numbering"))

	ch.canvasService.MustSaveCanvasEdits(r.Context(), workId, &service.CanvasEdits{
		Title:           extractOptionalFieldFromRequest(r, "title-text"),
		TextContents:    extractIndexedFieldsFromRequest(r, "text"),
		CaptionContents: extractIndexedFieldsFromRequest(r, "caption"),
	})
}

func (ch *CanvasHandler) handleAddText(rw http.ResponseWriter, r *http.Request, groupId int64) {
	ch.canvasService.MustInsertNewTextInGroup(r.Context(), groupId)
	ch.handleGet(rw, r, true)
}

func (ch *CanvasHandler) handleAddMedia(rw http.ResponseWriter, r *http.Request, groupId int64) {
	ch.canvasService.MustInsertNewMediaInGroup(r.Context(), groupId)
	ch.handleGet(rw, r, true)
}

func (ch *CanvasHandler) handleAddGroup(rw http.ResponseWriter, r *http.Request, groupId int64) {
	ch.canvasService.MustInsertNewGroupInGroup(r.Context(), groupId)
	ch.handleGet(rw, r, true)
}

func (ch *CanvasHandler) handleAddCaption(rw http.ResponseWriter, r *http.Request, mediaId int64) {
	ch.canvasService.MustInsertNewCaptionInMedia(r.Context(), mediaId)
	ch.handleGet(rw, r, true)
}

func (ch *CanvasHandler) handleContentIncreasePosition(rw http.ResponseWriter, r *http.Request, contentId int64) {
	ch.canvasService.MustIncreaseContentPosition(r.Context(), contentId)
	ch.handleGet(rw, r, true)
}

func (ch *CanvasHandler) handleContentDecreasePosition(rw http.ResponseWriter, r *http.Request, contentId int64) {
	ch.canvasService.MustDecreaseContentPosition(r.Context(), contentId)
	ch.handleGet(rw, r, true)
}

func (ch *CanvasHandler) handleDeleteContent(rw http.ResponseWriter, r *http.Request, contentId int64) {
	ch.canvasService.MustDeleteContent(r.Context(), contentId)
	ch.handleGet(rw, r, true)
}

func (ch *CanvasHandler) handleDeleteCaption(rw http.ResponseWriter, r *http.Request, captionId int64) {
	ch.canvasService.MustDeleteCaption(r.Context(), captionId)
	ch.handleGet(rw, r, true)
}

func (ch *CanvasHandler) handleUpload(rw http.ResponseWriter, r *http.Request, mediaId int64) {
	ch.canvasService.MustUploadMedia(r.Context(), mediaId, r)
	ch.handleGet(rw, r, true)
}
