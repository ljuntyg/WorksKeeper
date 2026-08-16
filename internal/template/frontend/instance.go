package frontend

import (
	"WorksKeeper/internal/repository"
	"fmt"
	"html/template"
)

type TemplateInstance struct {
	Instance           *repository.Instance
	TemplateCollection Templatable
}

func (ti *TemplateInstance) GetName(prefix string) string {
	return prefix + "instance"
}

func (ti *TemplateInstance) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/container/%s.html", fileName)
}

func (ti *TemplateInstance) ToEditableHtml() template.HTML {
	return mustToEditableHtml(ti)
}

func (ti *TemplateInstance) ToHtml() template.HTML {
	return mustToHtml(ti)
}
