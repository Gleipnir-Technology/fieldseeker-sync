package main

import (
	"bytes"
	"net/http"

	"gleipnir.technology/fieldseeker-sync-bridge"
)

type PageDataIndex struct {
	ServiceRequestCount int
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	count, err := fssync.ServiceRequestCount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageDataIndex{
		ServiceRequestCount: count,
	}

	// Create a buffer to hold the rendered template
	var buf bytes.Buffer

	// Execute template into buffer instead of directly to ResponseWriter
	err = tmpl.ExecuteTemplate(&buf, "base.html", data)
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
