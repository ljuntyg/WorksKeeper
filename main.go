package main

import (
	"WorksKeeper/internal/handler"
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/service"
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()

	// TODO: using mux to handle means .jpg, . will get replaced with / from subdomainPeriodReplacer
	/* for _, mt := range AllMediaTypes {
		mediaDir := "." + mt.mediaDir() // ./{resourcefolder}/{media}/{contentType (plural)}/
		urlPrefix := mt.urlPrefix()     // /{contentType (singular)}/{with}/{name}/{fileName}/
		log.Printf("mediaDir: %s, urlPrefix: %s", mediaDir, urlPrefix)

		contentTypeFileServer := http.FileServer(http.Dir(mediaDir))
		mux.Handle(urlPrefix, http.StripPrefix(urlPrefix, contentTypeFileServer))
	} */

	/* cssFileServer := http.FileServer(http.Dir("./resources/static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", cssFileServer)) */

	pgxPool := repository.GetPgxPool(
		os.Getenv("PGHOST"),
		os.Getenv("PGDATABASE"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
		os.Getenv("PGPORT"),
	)

	canvasRepo := &repository.CanvasRepository{}
	canvasRepo.Init(pgxPool)

	captionRepo := &repository.CaptionRepository{}
	captionRepo.Init(pgxPool)

	groupRepo := &repository.GroupRepository{}
	groupRepo.Init(pgxPool)

	mediaRepo := &repository.MediaRepository{}
	mediaRepo.Init(pgxPool)

	sourceRepo := &repository.SourceRepository{}
	sourceRepo.Init(pgxPool)

	textRepo := &repository.TextRepository{}
	textRepo.Init(pgxPool)

	workRepo := &repository.WorkRepository{}
	workRepo.Init(pgxPool)

	seriesRepo := &repository.SeriesRepository{}
	seriesRepo.Init(pgxPool)

	canvasService := &service.CanvasService{}
	canvasService.Init(canvasRepo, captionRepo, groupRepo, mediaRepo, sourceRepo, textRepo, workRepo)

	homeService := &service.HomeService{}
	homeService.Init(seriesRepo, workRepo)

	seriesService := &service.SeriesService{}
	seriesService.Init(seriesRepo, workRepo)

	workService := &service.WorkService{}
	workService.Init(canvasRepo, captionRepo, groupRepo, mediaRepo, sourceRepo, textRepo, workRepo)

	canvasHandler := &handler.CanvasHandler{}
	canvasHandler.Init(canvasService)

	homeHandler := &handler.HomeHandler{}
	homeHandler.Init(homeService)

	seriesHandler := &handler.SeriesHandler{}
	seriesHandler.Init(seriesService)

	workHandler := &handler.WorkHandler{}
	workHandler.Init(workService)

	mux.HandleFunc("/{$}", homeHandler.HandleRequest)
	mux.HandleFunc("/work/{numbering}", workHandler.HandleRequest)
	mux.HandleFunc("/series/{numbering}", seriesHandler.HandleRequest)
	mux.HandleFunc("/compose/work/{numbering}", canvasHandler.HandleRequest)
	mux.HandleFunc("/compose/work", canvasHandler.HandleRequestNew)

	/* mux.HandleFunc("/works", viewWorksHandler)

	mux.HandleFunc("/search/works", searchWorksHandler)

	mux.HandleFunc("/work/{numbering}", viewWorkHandler)
	mux.HandleFunc("/series/{numbering}", viewSeriesHandler)

	mux.HandleFunc("/compose/work", createNewWorkHandler)
	mux.HandleFunc("/compose/work/{numbering}", createWorkHandler)

	mux.HandleFunc("/organize/works/by/{name}", organizeWorksHandler) */

	// defering close in GetPgxPool causes "closed pool" error
	defer pgxPool.Close()

	log.Fatal(http.ListenAndServe(":8080", handler.SubdomainPeriodReplacer(mux)))
}
