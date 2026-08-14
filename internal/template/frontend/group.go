package frontend

import (
	"WorksKeeper/internal/repository"
	"fmt"
	"html/template"
)

type TemplateGroup struct {
	Group            *repository.Group
	TemplateContents []Templatable
}

func (tg *TemplateGroup) GetName(prefix string) string {
	return prefix + "group"
}

func (tg *TemplateGroup) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/content/%s.html", fileName)
}

func (tg *TemplateGroup) ToEditableHtml() template.HTML {
	return mustToEditableHtml(tg)
}

func (tg *TemplateGroup) ToHtml() template.HTML {
	return mustToHtml(tg)
}
