package handler

import (
	"WorksKeeper/internal/service"
	"log"
	"net/http"
	"strconv"
)

type EditSeriesHandler struct {
	editSeriesService *service.EditSeriesService
}

func (esh *EditSeriesHandler) Init(editSeriesService *service.EditSeriesService) {
	esh.editSeriesService = editSeriesService
}

func (esh *EditSeriesHandler) HandleRequest(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		esh.handleGet(rw, r, true)
	case http.MethodPost:
		esh.handlePost(rw, r)
	}
}

func (esh *EditSeriesHandler) HandleRequestNew(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		esh.handleGetNew(rw, r)
	}
}

func (esh *EditSeriesHandler) handleGet(rw http.ResponseWriter, r *http.Request, editing bool) {
	seriesId := mustNumberingStringToId(r.PathValue("numbering"))
	esh.editSeriesService.GetTemplateData(r.Context(), seriesId, editing).ExecuteTemplate(rw)
}

func (esh *EditSeriesHandler) handleGetNew(rw http.ResponseWriter, r *http.Request) {
	series := esh.editSeriesService.MustInsertNewSeriesInInstance(r.Context())
	http.Redirect(rw, r, "/compose"+series.GetNumberingUrlString(), http.StatusFound)
}

func (esh *EditSeriesHandler) handlePost(rw http.ResponseWriter, r *http.Request) {
	action, value := extractActionAndValueFromRequest(r)
	log.Println(action, value)
	esh.saveEdits(r)

	switch action {
	case "edit-series":
		esh.handleGet(rw, r, true)
	case "view-series":
		esh.handleGet(rw, r, false)
	case "add-series-work":
		seriesId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Series id")
		}

		esh.handleAddWork(rw, r, int64(seriesId))
	case "add-series-series":
		seriesId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Series id")
		}

		esh.handleAddSeries(rw, r, int64(seriesId))
	case "listing-increase-position":
		listingId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Listing id")
		}

		esh.handleListingIncreasePosition(rw, r, int64(listingId))
	case "listing-decrease-position":
		listingId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Listing id")
		}

		esh.handleListingDecreasePosition(rw, r, int64(listingId))
	case "delete-listing":
		listingId, err := strconv.Atoi(value)
		if err != nil {
			panic("unexpected error getting Listing id")
		}

		esh.handleDeleteListing(rw, r, int64(listingId))
	}
}

func (esh *EditSeriesHandler) saveEdits(r *http.Request) {
	seriesId := mustNumberingStringToId(r.PathValue("numbering"))
	esh.editSeriesService.MustSaveEditSeriesEdits(r.Context(), seriesId,
		extractOptionalFieldFromRequest(r, "title-text"))
}

func (esh *EditSeriesHandler) handleAddWork(rw http.ResponseWriter, r *http.Request, seriesId int64) {
	esh.editSeriesService.MustInsertNewWorkInSeries(r.Context(), seriesId)
	esh.handleGet(rw, r, true)
}

func (esh *EditSeriesHandler) handleAddSeries(rw http.ResponseWriter, r *http.Request, seriesId int64) {
	esh.editSeriesService.MustInsertNewSeriesInSeries(r.Context(), seriesId)
	esh.handleGet(rw, r, true)
}

func (esh *EditSeriesHandler) handleListingIncreasePosition(rw http.ResponseWriter, r *http.Request, listingId int64) {
	esh.editSeriesService.MustIncreaseListingPosition(r.Context(), listingId)
	esh.handleGet(rw, r, true)
}

func (esh *EditSeriesHandler) handleListingDecreasePosition(rw http.ResponseWriter, r *http.Request, listingId int64) {
	esh.editSeriesService.MustDecreaseListingPosition(r.Context(), listingId)
	esh.handleGet(rw, r, true)
}

func (esh *EditSeriesHandler) handleDeleteListing(rw http.ResponseWriter, r *http.Request, listingId int64) {
	esh.editSeriesService.MustDeleteListing(r.Context(), listingId)
	esh.handleGet(rw, r, true)
}
