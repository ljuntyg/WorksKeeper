package entity

import (
	"html/template"
	"log"
)

// TODO:
var WorkTemplate, EditableWorkTemplate *template.Template

func init() {
	WorkTemplate = LoadTemplate("resources/templates/listable/work.html")
	//EditableWorkTemplate = LoadTemplate("resources/templates/listable/editable/work.html")
}

func LoadTemplate(filePath string) *template.Template {
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		log.Fatal(err)
	}

	return tmpl
}
