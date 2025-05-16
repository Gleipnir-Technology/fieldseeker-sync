package main

import (
	"bytes"
	"net/http"

	"gleipnir.technology/fieldseeker-sync-bridge"
)

type PageDataIndex struct {
	ServiceRequestCount int
	Title               string
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	count, err := fssync.ServiceRequestCount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageDataIndex{
		ServiceRequestCount: count,
		Title:               "Gateway Sync Test",
	}

	renderTemplateOrError(w, r, "index.html", data)
}

func HandleServiceRequest(w http.ResponseWriter, r *http.Request) {
	requests, err := fssync.ServiceRequests()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderTemplateOrError(w, r, "service-requests.html", requests)
}

func renderTemplateOrError(w http.ResponseWriter, r *http.Request, template string, data any) {
	// Create a buffer to hold the rendered template
	var buf bytes.Buffer

	// Execute template into buffer instead of directly to ResponseWriter
	err := tmpl.ExecuteTemplate(&buf, template, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If template executed successfully, write the buffer to ResponseWriter
	_, err = buf.WriteTo(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
