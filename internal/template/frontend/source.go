package frontend

import (
	"WorksKeeper/internal/repository"
	"fmt"
	"html/template"
	"net"
	"net/url"
	"path"
	"strconv"
)

type TemplateSource struct {
	Source     *repository.Source
	Filename   *repository.Filename
	File       *repository.File
	Filenode   *repository.Filenode
	Fileserver *repository.Fileserver
}

// GetLink is the absolute url the stored file is served under, assembled from
// the Fileserver it lives on and the path it has on that Fileserver.
func (ts *TemplateSource) GetLink() string {
	link := url.URL{
		Scheme: ts.Fileserver.Scheme,
		Host:   ts.getHost(),
		Path:   path.Join(ts.Fileserver.UrlPath, ts.Filenode.Path),
	}

	return link.String()
}

// getHost leaves the port off when it is the default one for the scheme.
func (ts *TemplateSource) getHost() string {
	port := int(ts.Fileserver.Port)

	if (ts.Fileserver.Scheme == "http" && port == 80) ||
		(ts.Fileserver.Scheme == "https" && port == 443) {
		return ts.Fileserver.Host
	}

	return net.JoinHostPort(ts.Fileserver.Host, strconv.Itoa(port))
}

// GetMediaKind is the kind of html element this Source has to be rendered in.
func (ts *TemplateSource) GetMediaKind() MediaType {
	if mediaFormat, found := MimeToMediaFormat[ts.File.MimeType]; found {
		return mediaFormat.MediaType
	}

	return MediaEmptyType
}

// GetMimeType is the value of the type attribute of a source element.
func (ts *TemplateSource) GetMimeType() string {
	return ts.File.MimeType
}

// GetFileName is the name the file was uploaded under.
func (ts *TemplateSource) GetFileName() string {
	return ts.Filename.Name
}

func (ts *TemplateSource) GetName(prefix string) string {
	return prefix + "source"
}

func (ts *TemplateSource) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/content/%s.html", fileName)
}

func (ts *TemplateSource) ToEditableHtml() template.HTML {
	return mustToEditableHtml(ts)
}

func (ts *TemplateSource) ToHtml() template.HTML {
	return mustToHtml(ts)
}
