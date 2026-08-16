package main

import (
	"WorksKeeper/internal/handler"
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/service"
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	mux := http.NewServeMux()

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
	instanceRepo := &repository.InstanceRepository{}
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
		instanceRepo,
		listingRepo,
		mediaRepo,
		seriesRepo,
		sourceRepo,
		textRepo,
		workRepo)

	instancePort, err := strconv.Atoi(os.Getenv("INSTANCE_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	// The address the Instance is reached on, which is the one it redirects to.
	instance := service.MustGetOrInsertInstance(context.Background(), &repository.InstanceArguments{
		Scheme: os.Getenv("INSTANCE_SCHEME"),
		Host:   os.Getenv("INSTANCE_HOST"),
		Port:   int32(instancePort),
		Title:  os.Getenv("INSTANCE_TITLE"),
	}, repoCollection)

	canvasService := &service.CanvasService{}
	canvasService.Init(repoCollection, &instance)

	homeService := &service.HomeService{}
	homeService.Init(repoCollection, &instance)

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

	mux.HandleFunc("/{$}", func(rw http.ResponseWriter, r *http.Request) {
		http.Redirect(rw, r, "/works", http.StatusMovedPermanently)
	})

	mux.HandleFunc("/works", homeHandler.HandleRequest)
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

	log.Fatal(http.ListenAndServe(":8080", handler.SubdomainPeriodReplacer(instance.Scheme, instance.Authority(), mux)))
}
