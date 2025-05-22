package html

import (
	"embed"
	"html/template"
	"io"
	"os"

	"gleipnir.technology/fieldseeker-sync-bridge"
)

//go:embed templates/*
var files embed.FS
var (
	index           = newBuiltTemplate("templates/index.html")
	serviceRequests = newBuiltTemplate("templates/service-requests.html")
)

type BuiltTemplate struct {
	Path     string
	Template *template.Template
}
type PageDataIndex struct {
	ServiceRequestCount int
	Title               string
}

func (bt *BuiltTemplate) ExecuteTemplate(w io.Writer, t string, data any) error {
	if bt.Template == nil {
		return parseFromDisk(bt.Path).ExecuteTemplate(w, t, data)
	} else {
		return bt.Template.ExecuteTemplate(w, t, data)
	}
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

func newBuiltTemplate(path string) BuiltTemplate {
	full_path := "html/" + path
	_, err := os.Stat(full_path)
	if err == nil {
		return BuiltTemplate{
			Path:     full_path,
			Template: nil,
		}
	}
	return BuiltTemplate{
		Path:     path,
		Template: parseEmbedded(path),
	}
}

func parseEmbedded(file string) *template.Template {
	funcMap := template.FuncMap{
		"geocode": geocode,
	}
	return template.Must(
		template.New("templates/base.html").Funcs(funcMap).ParseFS(files, "templates/base.html", file))
}

func parseFromDisk(path string) *template.Template {
	funcMap := template.FuncMap{
		"geocode": geocode,
	}
	return template.Must(
		template.New("html/templates/base.html").Funcs(funcMap).ParseFiles(path, "html/templates/base.html", path))
}
