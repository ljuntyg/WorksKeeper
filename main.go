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

	// one file server per Fileserver row; disk_path is served under url_path,
	// except that only the stored files under it are, never the staging directory
	mediaFileserver := &repository.Fileserver{DiskPath: "./resources/media"}
	mediaFs := http.FileServer(http.Dir(mediaFileserver.ObjectsPath()))
	mux.Handle("/media/", http.StripPrefix("/media/", handler.WithoutContentSniffing(mediaFs)))

	pgxPool := repository.GetPgxPool(
		os.Getenv("PGHOST"),
		os.Getenv("PGDATABASE"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
		os.Getenv("PGPORT"),
	)

	canvasRepo := &repository.CanvasRepository{}
	captionRepo := &repository.CaptionRepository{}
	collectionRepo := &repository.CollectionRepository{}
	contentRepo := &repository.ContentRepository{}
	fileRepo := &repository.FileRepository{}
	filenameRepo := &repository.FilenameRepository{}
	filenodeRepo := &repository.FilenodeRepository{}
	fileserverRepo := &repository.FileserverRepository{}
	groupRepo := &repository.GroupRepository{}
	listingRepo := &repository.ListingRepository{}
	mediaRepo := &repository.MediaRepository{}
	sourceRepo := &repository.SourceRepository{}
	textRepo := &repository.TextRepository{}
	workRepo := &repository.WorkRepository{}
	seriesRepo := &repository.SeriesRepository{}

	repoCollection := &repository.RepositoryCollection{}
	repoCollection.Init(
		pgxPool,
		canvasRepo,
		captionRepo,
		collectionRepo,
		contentRepo,
		fileRepo,
		filenameRepo,
		filenodeRepo,
		fileserverRepo,
		groupRepo,
		listingRepo,
		mediaRepo,
		seriesRepo,
		sourceRepo,
		textRepo,
		workRepo)

	canvasService := &service.CanvasService{}
	canvasService.Init(repoCollection)

	homeService := &service.HomeService{}
	homeService.Init(repoCollection)

	seriesService := &service.SeriesService{}
	seriesService.Init(repoCollection)

	workService := &service.WorkService{}
	workService.Init(repoCollection)

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
