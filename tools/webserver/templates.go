package fssync

import (
	"embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed templates/*
var templateFiles embed.FS
var tmpl *template.Template

type PageData struct {
	Title   string
	Message string
}

func InitializeTemplates() {
	var err error
	tmpl, err = template.ParseFS(templateFiles, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

}
func HandleIndex(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Welcome",
		Message: "Hello from embedded templates!",
	}

	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// templates/base.html
/*
 */

// templates/home.html
/*
 */
