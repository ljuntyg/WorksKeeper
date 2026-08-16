package service

import (
	"WorksKeeper/internal/repository"
	"WorksKeeper/internal/template"
	"WorksKeeper/internal/template/frontend"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/jackc/pgx/v5"
)

type CanvasService struct {
	repos    *repository.RepositoryCollection
	instance *repository.Instance
}

func (cs *CanvasService) Init(repos *repository.RepositoryCollection, instance *repository.Instance) {
	cs.repos = repos
	cs.instance = instance
}

func (cs *CanvasService) GetTemplateData(ctx context.Context, workId int64, editing bool) template.Executable {
	templateWork := mustBuildTemplateWorkShallow(ctx, workId, cs.repos)
	mustFillTemplateWork(ctx, templateWork, cs.repos)

	return &template.CanvasData{
		TemplateWork: templateWork,
		IsEditing:    editing,
	}
}

func (cs *CanvasService) MustInsertNewWorkInBaseCollection(ctx context.Context) *frontend.TemplateWork {
	collection, err := cs.repos.CollectionRepo.GetOneCollectionByInstanceId(ctx, cs.instance.Id)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Collection")
	}

	rootSeries, err := cs.repos.SeriesRepo.GetOneSeriesByCollectionId(ctx, collection.Id)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Series")
	}

	return cs.mustInsertNewWork(ctx, rootSeries.Id)
}

func (cs *CanvasService) MustInsertNewTextInGroup(ctx context.Context, groupId int64) *frontend.TemplateText {
	tx := cs.repos.MustBegin(ctx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error inserting new Text; rolled back")
		}
	}()

	templateContent := mustInsertNewTemplateTextContent(ctx, tx, groupId, cs.repos)

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error committing new Text")
	}

	return templateContent.TemplateGroupOrTextOrMedia.(*frontend.TemplateText)
}

func (cs *CanvasService) MustInsertNewMediaInGroup(ctx context.Context, groupId int64) *frontend.TemplateMedia {
	tx := cs.repos.MustBegin(ctx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error inserting new Media; rolled back")
		}
	}()

	templateContent := mustInsertNewTemplateMediaContent(ctx, tx, groupId, cs.repos)

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error committing new Media")
	}

	return templateContent.TemplateGroupOrTextOrMedia.(*frontend.TemplateMedia)
}

func (cs *CanvasService) MustInsertNewGroupInGroup(ctx context.Context, groupId int64) *frontend.TemplateGroup {
	tx := cs.repos.MustBegin(ctx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error inserting new Group; rolled back")
		}
	}()

	templateContent := mustInsertNewTemplateGroupContent(ctx, tx, groupId, cs.repos)

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error committing new Group")
	}

	return templateContent.TemplateGroupOrTextOrMedia.(*frontend.TemplateGroup)
}

func (cs *CanvasService) MustInsertNewCaptionInMedia(ctx context.Context, mediaId int64) *frontend.TemplateCaption {
	tx := cs.repos.MustBegin(ctx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error inserting new Caption; rolled back")
		}
	}()

	templateCaption := mustInsertNewTemplateCaption(ctx, tx, mediaId, cs.repos)

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error committing new Caption")
	}

	return templateCaption
}

func (cs *CanvasService) MustIncreaseContentPosition(ctx context.Context, contentId int64) {
	tx := cs.repos.MustBegin(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error increasing Content position; rolled back")
		}
	}()

	if _, err := cs.repos.ContentRepo.IncreaseContentPositionTx(ctx, tx, contentId); err != nil {
		log.Println(err)
		panic("unexpected error increasing Content position")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error increasing Content position")
	}
}

func (cs *CanvasService) MustDecreaseContentPosition(ctx context.Context, contentId int64) {
	tx := cs.repos.MustBegin(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error decreasing Content position; rolled back")
		}
	}()

	if _, err := cs.repos.ContentRepo.DecreaseContentPositionTx(ctx, tx, contentId); err != nil {
		log.Println(err)
		panic("unexpected error decreasing Content position")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error decreasing Content position")
	}
}

func (cs *CanvasService) MustDeleteContent(ctx context.Context, contentId int64) {
	tx := cs.repos.MustBegin(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error deleting Content; rolled back")
		}
	}()

	if err := cs.repos.ContentRepo.DeleteContentByIdTx(ctx, tx, contentId); err != nil {
		log.Println(err)
		panic("unexpected error deleting Content")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error deleting Content")
	}
}

func (cs *CanvasService) MustDeleteCaption(ctx context.Context, captionId int64) {
	tx := cs.repos.MustBegin(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error deleting Caption; rolled back")
		}
	}()

	if err := cs.repos.CaptionRepo.DeleteCaptionByIdTx(ctx, tx, captionId); err != nil {
		log.Println(err)
		panic("unexpected error deleting Caption")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error deleting Caption")
	}
}

type CanvasEdits struct {
	Title           *string
	TextContents    map[int64]string
	CaptionContents map[int64]string
}

func (cs *CanvasService) MustSaveCanvasEdits(ctx context.Context, workId int64, edits *CanvasEdits) {
	if edits.Title == nil && len(edits.TextContents) == 0 && len(edits.CaptionContents) == 0 {
		return
	}

	tx := cs.repos.MustBegin(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error saving Canvas edits; rolled back")
		}
	}()

	if edits.Title != nil {
		if _, err := cs.repos.WorkRepo.UpdateWorkSetTitleByIdTx(ctx, tx, workId, *edits.Title); err != nil {
			log.Println(err)
			panic("unexpected error saving Work title")
		}
	}

	for textId, content := range edits.TextContents {
		if _, err := cs.repos.TextRepo.UpdateTextSetContentByIdTx(ctx, tx, textId, content); err != nil {
			log.Println(err)
			panic("unexpected error saving Text content")
		}
	}

	for captionId, content := range edits.CaptionContents {
		if _, err := cs.repos.CaptionRepo.UpdateCaptionSetContentByIdTx(ctx, tx, captionId, content); err != nil {
			log.Println(err)
			panic("unexpected error saving Caption content")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error saving Canvas edits")
	}
}

func (cs *CanvasService) MustUploadMedia(ctx context.Context, mediaId int64, r *http.Request) {
	fileserver, err := cs.repos.FileserverRepo.GetOneFileserverOrderByIdAscending(ctx)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Fileserver to upload to")
	}

	upload := mustStoreUploadedFile(r, &fileserver)

	tx := cs.repos.MustBegin(ctx)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error uploading Media; rolled back")
		}
	}()

	file := mustGetOrInsertFile(ctx, tx, upload, cs.repos)
	mustGetOrInsertFilenode(ctx, tx, file.Id, fileserver.Id, upload.path, cs.repos)
	filename := mustGetOrInsertFilename(ctx, tx, file.Id, upload.name, cs.repos)

	if _, err := cs.repos.SourceRepo.InsertSourceTx(ctx, tx, &repository.SourceArguments{
		MediaId:    mediaId,
		FilenameId: filename.Id,
	}); err != nil {
		log.Println(err)
		panic("unexpected error inserting new Source")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error uploading Media")
	}
}

func (cs *CanvasService) mustInsertNewWork(ctx context.Context, seriesId int64) *frontend.TemplateWork {
	tx := cs.repos.MustBegin(ctx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(context.WithoutCancel(ctx))
			log.Println(r)
			panic("unexpected error inserting new Work; rolled back")
		}
	}()

	templateListing := mustInsertNewTemplateWorkListing(ctx, tx, seriesId, cs.repos)

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		panic("unexpected error committing new Work")
	}

	return templateListing.TemplateWorkOrSeries.(*frontend.TemplateWork)
}

func mustGetOrInsertFile(ctx context.Context, tx pgx.Tx, upload *uploadedFile, repos *repository.RepositoryCollection) repository.File {
	file, err := repos.FileRepo.GetOptionalFileByHashAndSizeTx(ctx, tx, upload.hash, upload.size)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting File by contents")
	}

	if file != nil {
		return *file
	}

	insertedFile, err := repos.FileRepo.InsertFileTx(ctx, tx, &repository.FileArguments{
		Size:     upload.size,
		Hash:     upload.hash,
		MimeType: upload.mimeType,
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new File")
	}

	return insertedFile
}

func mustGetOrInsertFilenode(ctx context.Context, tx pgx.Tx, fileId int64, fileserverId int64, path string, repos *repository.RepositoryCollection) repository.Filenode {
	filenode, err := repos.FilenodeRepo.GetOptionalFilenodeByFileserverIdAndPathTx(ctx, tx, fileserverId, path)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Filenode by location")
	}

	if filenode != nil {
		return *filenode
	}

	insertedFilenode, err := repos.FilenodeRepo.InsertFilenodeTx(ctx, tx, &repository.FilenodeArguments{
		FileId:       fileId,
		FileserverId: fileserverId,
		Path:         path,
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Filenode")
	}

	return insertedFilenode
}

func mustGetOrInsertFilename(ctx context.Context, tx pgx.Tx, fileId int64, name string, repos *repository.RepositoryCollection) repository.Filename {
	filename, err := repos.FilenameRepo.GetOptionalFilenameByFileIdAndNameTx(ctx, tx, fileId, name)
	if err != nil {
		log.Println(err)
		panic("unexpected error getting Filename by name")
	}

	if filename != nil {
		return *filename
	}

	insertedFilename, err := repos.FilenameRepo.InsertFilenameTx(ctx, tx, &repository.FilenameArguments{
		FileId: fileId,
		Name:   name,
	})

	if err != nil {
		log.Println(err)
		panic("unexpected error inserting new Filename")
	}

	return insertedFilename
}

type uploadedFile struct {
	name     string
	path     string
	hash     string
	size     int64
	mimeType string
}

func mustStoreUploadedFile(r *http.Request, fileserver *repository.Fileserver) *uploadedFile {
	file, fileHeader, err := r.FormFile("upload")
	if err != nil {
		log.Println(err)
		panic("unexpected error getting uploaded file")
	}
	defer file.Close()

	mimeType, extension := mustResolveUploadFormat(fileHeader)
	hash, size := mustHashUploadedFile(file)
	path := path.Join(hash[0:2], hash[2:4], hash+extension)

	upload := &uploadedFile{
		name:     fileHeader.Filename,
		path:     path,
		hash:     hash,
		size:     size,
		mimeType: mimeType,
	}

	mustWriteUploadedFile(
		file,
		filepath.Join(fileserver.ObjectsPath(), filepath.FromSlash(upload.path)),
		fileserver.TempPath(),
		size,
	)

	return upload
}

func mustResolveUploadFormat(fileHeader *multipart.FileHeader) (string, string) {
	mimeType := fileHeader.Header.Get("Content-Type")

	mediaFormat, found := frontend.MimeToMediaFormat[mimeType]
	if !found {
		log.Printf("unsupported media type %q", mimeType)
		panic("unsupported media type")
	}

	return mimeType, mediaFormat.Extension
}

func mustHashUploadedFile(file multipart.File) (string, int64) {
	hash := sha256.New()

	size, err := io.Copy(hash, file)
	if err != nil {
		log.Println(err)
		panic("unexpected error reading uploaded file")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Println(err)
		panic("unexpected error rewinding uploaded file")
	}

	return hex.EncodeToString(hash.Sum(nil)), size
}

// mustWriteUploadedFile writes the uploaded file to the name its contents give
// it, unless we are already holding those contents there.
func mustWriteUploadedFile(file multipart.File, storedPath string, tempDir string, size int64) {
	if err := os.MkdirAll(filepath.Dir(storedPath), 0o755); err != nil {
		log.Println(err)
		panic("unexpected error creating media directory")
	}

	if storedFile, err := os.Stat(storedPath); err == nil && storedFile.Size() == size {
		return
	}

	tempPath := mustWriteTempFile(tempDir, file)

	if err := os.Rename(tempPath, storedPath); err != nil {
		os.Remove(tempPath)
		log.Println(err)
		panic("unexpected error storing uploaded file")
	}
}

// mustWriteTempFile writes the uploaded file to the staging directory, which is
// not served and which shares a file system with where the file is to be
// stored, so that it is renamed there rather than copied.
func mustWriteTempFile(dir string, file multipart.File) string {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Println(err)
		panic("unexpected error creating staging directory")
	}

	tempFile, err := os.CreateTemp(dir, "upload-*")
	if err != nil {
		log.Println(err)
		panic("unexpected error creating temporary file")
	}
	defer tempFile.Close()

	if err := tempFile.Chmod(0o644); err != nil {
		os.Remove(tempFile.Name())
		log.Println(err)
		panic("unexpected error setting uploaded file permissions")
	}

	if _, err := io.Copy(tempFile, file); err != nil {
		os.Remove(tempFile.Name())
		log.Println(err)
		panic("unexpected error writing uploaded file")
	}

	if err := tempFile.Sync(); err != nil {
		os.Remove(tempFile.Name())
		log.Println(err)
		panic("unexpected error flushing uploaded file")
	}

	return tempFile.Name()
}
