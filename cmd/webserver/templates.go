package main

import (
	"embed"
	"html/template"
	"log"
)

//go:embed templates/*
var templateFiles embed.FS
var tmpl *template.Template

func InitializeTemplates() {
	var err error
	tmpl, err = template.ParseFS(templateFiles, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

}
