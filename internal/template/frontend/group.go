package frontend

import (
	"WorksKeeper/internal/repository/entity"
	"fmt"
	"html/template"
)

type TemplateGroup struct {
	Group        *entity.Group
	Templatables []Templatable
}

func (tg *TemplateGroup) GetName(prefix string) string {
	return tg.Group.GetName(prefix)
}

func (tg *TemplateGroup) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/content/groups/%s.html", fileName)
}

func (tg *TemplateGroup) ToEditableHtml() template.HTML {
	return mustToEditableHtml(tg)
}

func (tg *TemplateGroup) ToHtml() template.HTML {
	return mustToHtml(tg)
}
