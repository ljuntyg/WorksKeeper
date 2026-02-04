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

	workRepo := &repository.WorkRepository{}
	workRepo.Init(pgxPool)

	seriesRepo := &repository.SeriesRepository{}
	seriesRepo.Init(pgxPool)

	homeService := &service.HomeService{}
	homeService.Init(workRepo, seriesRepo)

	homeHandler := &handler.HomeHandler{}
	homeHandler.Init(homeService)

	mux.HandleFunc("/{$}", homeHandler.HandleRequest)
	/* mux.HandleFunc("/works", viewWorksHandler)

	mux.HandleFunc("/search/works", searchWorksHandler)

	mux.HandleFunc("/work/{numbering}", viewWorkHandler)
	mux.HandleFunc("/series/{numbering}", viewSeriesHandler)

	mux.HandleFunc("/compose/work", createNewWorkHandler)
	mux.HandleFunc("/compose/work/{numbering}", createWorkHandler)

	mux.HandleFunc("/organize/works/by/{name}", organizeWorksHandler) */

	log.Fatal(http.ListenAndServe(":8080", handler.SubdomainPeriodReplacer(mux)))
}
