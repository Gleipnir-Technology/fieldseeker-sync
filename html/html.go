package html

import (
	"embed"
	"html/template"
	"io"

	"gleipnir.technology/fieldseeker-sync-bridge"
)

//go:embed templates/*
var files embed.FS
var (
	index           = parse("templates/index.html")
	serviceRequests = parse("templates/service-requests.html")
)

type PageDataIndex struct {
	ServiceRequestCount int
	Title               string
}

func Index(w io.Writer, d PageDataIndex) error {
	return index.ExecuteTemplate(w, "index.html", d)
}

func ServiceRequests(w io.Writer, sr []*fssync.ServiceRequest) error {
	return serviceRequests.ExecuteTemplate(w, "service-requests.html", sr)
}

func geocode(geo fssync.Geometry) string {
	return "foo"
}

func parse(file string) *template.Template {
	funcMap := template.FuncMap{
		"geocode": geocode,
	}
	return template.Must(
		template.New("templates/base.html").Funcs(funcMap).ParseFS(files, "templates/base.html", file))
}
