package handler

import (
	"WorksKeeper/internal/service"
	"log"
	"net/http"
	"strconv"
)

type EditWorkHandler struct {
	editWorkService *service.EditWorkService
}

// maxUploadSize is the largest body a POST is read from. Without a limit a
// single request can fill the disk, because the parts of a multipart form that
// do not fit in memory are spilled to a temporary file however large they are.
const maxUploadSize = 256 << 20

func (ewh *EditWorkHandler) Init(editWorkService *service.EditWorkService) {
	ewh.editWorkService = editWorkService
}

func (ewh *EditWorkHandler) HandleRequest(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ewh.handleGet(rw, r, true)
	case http.MethodPost:
		ewh.handlePost(rw, r)
	}
}

func (ewh *EditWorkHandler) HandleRequestNew(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ewh.handleGetNew(rw, r)
	}
}

func (ewh *EditWorkHandler) handleGet(rw http.ResponseWriter, r *http.Request, editing bool) {
	workId := mustNumberingStringToId(r.PathValue("numbering"))
	ewh.editWorkService.GetTemplateData(r.Context(), workId, editing).ExecuteTemplate(rw)
}

func (ewh *EditWorkHandler) handlePost(rw http.ResponseWriter, r *http.Request) {
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
	ewh.saveEdits(r)

	switch action {
	case "edit-work":
		ewh.handleGet(rw, r, true)
	case "view-work":
		ewh.handleGet(rw, r, false)
	case "add-text":
		groupId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Group id")
		}

		ewh.handleAddText(rw, r, int64(groupId))
	case "add-work-text":
		workId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Work id")
		}

		ewh.handleAddTextToWork(rw, r, int64(workId))
	case "add-media":
		groupId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Group id")
		}

		ewh.handleAddMedia(rw, r, int64(groupId))
	case "add-work-media":
		workId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Work id")
		}

		ewh.handleAddMediaToWork(rw, r, int64(workId))
	case "add-group":
		groupId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Group id")
		}

		ewh.handleAddGroup(rw, r, int64(groupId))
	case "add-work-group":
		workId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Work id")
		}

		ewh.handleAddGroupToWork(rw, r, int64(workId))
	case "set-media-caption":
		mediaId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Media id")
		}

		ewh.handleSetMediaCaption(rw, r, int64(mediaId))
	case "content-increase-position":
		contentId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Content id")
		}

		ewh.handleContentIncreasePosition(rw, r, int64(contentId))
	case "content-decrease-position":
		contentId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Content id")
		}

		ewh.handleContentDecreasePosition(rw, r, int64(contentId))
	case "delete-content":
		contentId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Content id")
		}

		ewh.handleDeleteContent(rw, r, int64(contentId))
	case "clear-media-caption":
		mediaId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Media id")
		}

		ewh.handleClearMediaCaption(rw, r, int64(mediaId))
	case "upload":
		mediaId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Media id")
		}

		ewh.handleUpload(rw, r, int64(mediaId))
	}
}

func (ewh *EditWorkHandler) handleGetNew(rw http.ResponseWriter, r *http.Request) {
	work := ewh.editWorkService.MustInsertNewWorkInInstance(r.Context())
	http.Redirect(rw, r, "/compose"+work.GetNumberingUrlString(), http.StatusFound)
}

func (ewh *EditWorkHandler) saveEdits(r *http.Request) {
	workId := mustNumberingStringToId(r.PathValue("numbering"))

	ewh.editWorkService.MustSaveEditWorkEdits(r.Context(), workId, &service.EditWorkEdits{
		Title:           extractOptionalFieldFromRequest(r, "title-text"),
		TextContents:    extractIndexedFieldsFromRequest(r, "text"),
		CaptionContents: extractIndexedFieldsFromRequest(r, "caption"),
	})
}

func (ewh *EditWorkHandler) handleAddText(rw http.ResponseWriter, r *http.Request, groupId int64) {
	ewh.editWorkService.MustInsertNewTextInGroup(r.Context(), groupId)
	ewh.handleGet(rw, r, true)
}

func (ewh *EditWorkHandler) handleAddTextToWork(rw http.ResponseWriter, r *http.Request, workId int64) {
	ewh.editWorkService.MustInsertNewTextInWork(r.Context(), workId)
	ewh.handleGet(rw, r, true)
}

func (ewh *EditWorkHandler) handleAddMedia(rw http.ResponseWriter, r *http.Request, groupId int64) {
	ewh.editWorkService.MustInsertNewMediaInGroup(r.Context(), groupId)
	ewh.handleGet(rw, r, true)
}

func (ewh *EditWorkHandler) handleAddMediaToWork(rw http.ResponseWriter, r *http.Request, workId int64) {
	ewh.editWorkService.MustInsertNewMediaInWork(r.Context(), workId)
	ewh.handleGet(rw, r, true)
}

func (ewh *EditWorkHandler) handleAddGroup(rw http.ResponseWriter, r *http.Request, groupId int64) {
	ewh.editWorkService.MustInsertNewGroupInGroup(r.Context(), groupId)
	ewh.handleGet(rw, r, true)
}

func (ewh *EditWorkHandler) handleAddGroupToWork(rw http.ResponseWriter, r *http.Request, workId int64) {
	ewh.editWorkService.MustInsertNewGroupInWork(r.Context(), workId)
	ewh.handleGet(rw, r, true)
}

func (ewh *EditWorkHandler) handleSetMediaCaption(rw http.ResponseWriter, r *http.Request, mediaId int64) {
	ewh.editWorkService.MustSetMediaCaption(r.Context(), mediaId, "")
	ewh.handleGet(rw, r, true)
}

func (ewh *EditWorkHandler) handleContentIncreasePosition(rw http.ResponseWriter, r *http.Request, contentId int64) {
	ewh.editWorkService.MustIncreaseContentPosition(r.Context(), contentId)
	ewh.handleGet(rw, r, true)
}

func (ewh *EditWorkHandler) handleContentDecreasePosition(rw http.ResponseWriter, r *http.Request, contentId int64) {
	ewh.editWorkService.MustDecreaseContentPosition(r.Context(), contentId)
	ewh.handleGet(rw, r, true)
}

func (ewh *EditWorkHandler) handleDeleteContent(rw http.ResponseWriter, r *http.Request, contentId int64) {
	ewh.editWorkService.MustDeleteContent(r.Context(), contentId)
	ewh.handleGet(rw, r, true)
}

func (ewh *EditWorkHandler) handleClearMediaCaption(rw http.ResponseWriter, r *http.Request, mediaId int64) {
	ewh.editWorkService.MustClearMediaCaption(r.Context(), mediaId)
	ewh.handleGet(rw, r, true)
}

func (ewh *EditWorkHandler) handleUpload(rw http.ResponseWriter, r *http.Request, mediaId int64) {
	ewh.editWorkService.MustUploadMedia(r.Context(), mediaId, r)
	ewh.handleGet(rw, r, true)
}
