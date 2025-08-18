package html

import (
	"embed"
	"errors"
	"html/template"
	"io"
	"log"
	"os"

	"github.com/Gleipnir-Technology/fieldseeker-sync/shared"
)

//go:embed templates/*
var embeddedFiles embed.FS
var (
	index           = newBuiltTemplate("index", "base")
	login           = newBuiltTemplate("login", "base")
	serviceRequests = newBuiltTemplate("service-requests", "base")
)
var components = [...]string{"navbar"}

type BuiltTemplate struct {
	files    []string
	template *template.Template
}

type PageDataIndex struct {
	ServiceRequestCount int
	Title               string
	User                *fssync.User
}

type PageDataLogin struct {
	Title string
}

type PageDataServiceRequests struct {
	ServiceRequests []fssync.ServiceRequest
	User            *fssync.User
}

func (bt *BuiltTemplate) ExecuteTemplate(w io.Writer, data any) error {
	name := bt.files[0] + ".html"
	if bt.template == nil {
		templ := parseFromDisk(bt.files)
		if templ == nil {
			w.Write([]byte("Failure."))
			return errors.New("Template parsing failed")
		}
		return templ.ExecuteTemplate(w, name, data)
	} else {
		return bt.template.ExecuteTemplate(w, name, data)
	}
}
func Index(w io.Writer, d PageDataIndex) error {
	return index.ExecuteTemplate(w, d)
}

func Login(w io.Writer) error {
	d := PageDataIndex{
		ServiceRequestCount: 0,
		Title:               "Login",
	}
	return login.ExecuteTemplate(w, d)
}

func ServiceRequests(w io.Writer, sr PageDataServiceRequests) error {
	return serviceRequests.ExecuteTemplate(w, sr)
}

func geocode(geo fssync.LatLong) string {
	return "foo"
}

func newBuiltTemplate(files ...string) BuiltTemplate {
	// If we are in dev mode we can tell because all the files we want
	// are available on disk and we should pull from them.
	files_on_disk := true
	for _, f := range files {
		full_path := "html/templates/" + f + ".html"
		_, err := os.Stat(full_path)
		if err != nil {
			files_on_disk = false
			break
		}
	}
	if files_on_disk {
		return BuiltTemplate{
			files:    files,
			template: nil,
		}
	}
	// If we are in production mode parse all the templates now
	return BuiltTemplate{
		files:    files,
		template: parseEmbedded(files),
	}
}

func parseEmbedded(files []string) *template.Template {
	funcMap := template.FuncMap{
		"geocode": geocode,
	}
	// Remap the file names to embedded paths
	paths := make([]string, 0)
	for _, f := range files {
		paths = append(paths, "templates/"+f+".html")
	}
	for _, f := range components {
		paths = append(paths, "templates/components/"+f+".html")
	}
	name := files[0]
	return template.Must(
		template.New(name).Funcs(funcMap).ParseFS(embeddedFiles, paths...))
}

func parseFromDisk(files []string) *template.Template {
	funcMap := template.FuncMap{
		"geocode": geocode,
	}
	// Remap file names to paths on disk
	paths := make([]string, 0)
	for _, f := range files {
		paths = append(paths, "html/templates/"+f+".html")
	}
	name := files[0] + ".html"
	for _, f := range components {
		paths = append(paths, "html/templates/components/"+f+".html")
	}
	templ, err := template.New(name).Funcs(funcMap).ParseFiles(paths...)
	if err != nil {
		log.Println("TEMPLATE FAILED", err)
		return nil
	}
	return templ
}
