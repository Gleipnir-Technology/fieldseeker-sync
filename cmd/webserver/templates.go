package main

import (
	"embed"
	"html/template"
	"log"

	"gleipnir.technology/fieldseeker-sync-bridge"
)

//go:embed templates/*
var templateFiles embed.FS
var tmpl *template.Template

func InitializeTemplates() {
	var err error
	funcMap := template.FuncMap{
		"geocode": geocode,
	}
	tmpl, err = template.New("root").Funcs(funcMap).ParseFS(templateFiles, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
}

func geocode(geo fssync.Geometry) string {
	return "foo"
}
