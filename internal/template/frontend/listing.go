package frontend

import (
	"WorksKeeper/internal/repository"
	"fmt"
	"html/template"
)

type TemplateListing struct {
	Listing              *repository.Listing
	TemplateWorkOrSeries Listable
}

func (tl *TemplateListing) GetName(prefix string) string {
	return prefix + "listing"
}

func (tl *TemplateListing) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/listing/%s.html", fileName)
}

func (tl *TemplateListing) ToEditableHtml() template.HTML {
	return mustToEditableHtml(tl)
}

func (tl *TemplateListing) ToHtml() template.HTML {
	return mustToHtml(tl)
}
