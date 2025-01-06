package render

import (
	"fmt"
	"html/template"
	"net/http"
)

var pathToTemplates = "./templates/"

// RenderTemplate renders templates using html/template
func RenderTemplate(w http.ResponseWriter, tmpl string) {
	parsedTemplate, _ := template.ParseFiles(pathToTemplates + tmpl)
	err := parsedTemplate.Execute(w, nil)
	if err != nil {
		fmt.Println("error parsing template: ", err)
		return
	}
}
