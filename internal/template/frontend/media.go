package frontend

import (
	"WorksKeeper/internal/repository"
	"fmt"
	"html/template"
)

// MediaType is the kind of html element a file has to be rendered in.
type MediaType string

const (
	MediaEmptyType MediaType = "empty"
	MediaImageType MediaType = "image"
	MediaSoundType MediaType = "sound"
	MediaVideoType MediaType = "video"
)

// MediaFormat is what we know about an accepted MIME type: how a file of that
// type has to be rendered, and the extension it is stored under so that it is
// served back as the type it was uploaded as.
type MediaFormat struct {
	MediaType MediaType
	Extension string
}

// MimeToMediaFormat maps the MIME types a browser sends in the Content-Type
// header of a multipart form file part to the formats we store. These are the
// MIME types browsers report, not the ones http.DetectContentType sniffs, and
// a type that is not in here is one we refuse to store.
//
// Nothing a browser can run a script from belongs in here. Stored files are
// served from the same origin as the pages that show them, so a format that
// carries code, image/svg+xml above all, would be a way to run that code as us.
var MimeToMediaFormat = map[string]MediaFormat{
	"image/apng":      {MediaImageType, ".apng"},
	"image/avif":      {MediaImageType, ".avif"},
	"image/bmp":       {MediaImageType, ".bmp"},
	"image/gif":       {MediaImageType, ".gif"},
	"image/jpeg":      {MediaImageType, ".jpg"},
	"image/png":       {MediaImageType, ".png"},
	"image/tiff":      {MediaImageType, ".tiff"},
	"image/webp":      {MediaImageType, ".webp"},
	"audio/aac":       {MediaSoundType, ".aac"},
	"audio/midi":      {MediaSoundType, ".midi"},
	"audio/x-midi":    {MediaSoundType, ".midi"},
	"audio/mpeg":      {MediaSoundType, ".mp3"},
	"audio/ogg":       {MediaSoundType, ".oga"},
	"audio/wav":       {MediaSoundType, ".wav"},
	"audio/webm":      {MediaSoundType, ".weba"},
	"audio/3gpp":      {MediaSoundType, ".3gp"},
	"audio/3gpp2":     {MediaSoundType, ".3g2"},
	"application/ogg": {MediaSoundType, ".ogx"}, // TODO: ? some .ogg files
	"video/mp4":       {MediaVideoType, ".mp4"},
	"video/mpeg":      {MediaVideoType, ".mpeg"},
	"video/ogg":       {MediaVideoType, ".ogv"},
	"video/webm":      {MediaVideoType, ".webm"},
	"video/x-msvideo": {MediaVideoType, ".avi"},
	"video/mp2t":      {MediaVideoType, ".ts"},
	"video/3gpp":      {MediaVideoType, ".3gp"},
	"video/3gpp2":     {MediaVideoType, ".3g2"},
}

type TemplateMedia struct {
	Media           *repository.Media
	TemplateSources []Templatable
	TemplateCaption Templatable
}

// GetMediaKind is the kind of html element this Media has to be rendered in.
// It is decided by the first Source, because a Media renders as a single
// element that all of its Sources are alternative encodings for.
func (tm *TemplateMedia) GetMediaKind() MediaType {
	templateSource := tm.firstTemplateSource()
	if templateSource == nil {
		return MediaEmptyType
	}

	return templateSource.GetMediaKind()
}

// GetFallbackLink is the url of the first Source. A picture element needs an
// img element as the fallback its source elements are alternatives to.
func (tm *TemplateMedia) GetFallbackLink() string {
	templateSource := tm.firstTemplateSource()
	if templateSource == nil {
		return ""
	}

	return templateSource.GetLink()
}

// firstTemplateSource narrows the first Source to its concrete type, which the
// templates cannot do themselves; a method that is not part of Templatable
// cannot be called on an element of TemplateSources from a template.
func (tm *TemplateMedia) firstTemplateSource() *TemplateSource {
	if len(tm.TemplateSources) == 0 {
		return nil
	}

	templateSource, ok := tm.TemplateSources[0].(*TemplateSource)
	if !ok {
		return nil
	}

	return templateSource
}

func (tm *TemplateMedia) GetName(prefix string) string {
	return prefix + "media"
}

func (tm *TemplateMedia) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/content/%s.html", fileName)
}

func (tm *TemplateMedia) ToEditableHtml() template.HTML {
	return mustToEditableHtml(tm)
}

func (tm *TemplateMedia) ToHtml() template.HTML {
	return mustToHtml(tm)
}
