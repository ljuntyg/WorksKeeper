package template

import (
	"html/template"
	"log"
)

func MustLoadTemplate(filePath string) *template.Template {
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		log.Fatal(err)
	}

	return tmpl
}
