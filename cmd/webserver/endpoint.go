package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	"github.com/Gleipnir-Technology/fieldseeker-sync"
	"github.com/Gleipnir-Technology/fieldseeker-sync/database"
	"github.com/Gleipnir-Technology/fieldseeker-sync/html"
	"github.com/Gleipnir-Technology/fieldseeker-sync/shared"
)

func apiAudioPost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	id := chi.URLParam(r, "uuid")
	noteUUID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Failed to decode the uuid", http.StatusBadRequest)
		return
	}

	var payload shared.NoteAudioPayload
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read the payload", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Println("Audio note POST JSON decode error: ", err)
		output, err := os.OpenFile("/tmp/request.body", os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Println("Failed to open temp request.bady")
		}
		defer output.Close()
		output.Write(body)
		log.Println("Wrote request to /tmp/request.body")

		http.Error(w, "Failed to decode the payload", http.StatusBadRequest)
		return
	}
	if err := database.NoteAudioCreate(context.Background(), noteUUID, payload, u.ID); err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func apiAudioContentPost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	u_str := chi.URLParam(r, "uuid")
	audioUUID, err := uuid.Parse(u_str)
	if err != nil {
		http.Error(w, "Failed to parse image UUID", http.StatusBadRequest)
		return
	}

	config, err := fssync.ReadConfig()
	if err != nil {
		log.Printf("Failed to read config", err)
		http.Error(w, "Unable to create file", http.StatusInternalServerError)
		return
	}
	filepath := fmt.Sprintf("%s/%s.m4a", config.UserFiles.Directory, audioUUID.String())

	// Create file in configured directory
	dst, err := os.Create(filepath)
	if err != nil {
		log.Printf("Failed to create audio file at %s: %v\n", dst, err)
		http.Error(w, "Unable to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy rest of request body to file
	_, err = io.Copy(dst, r.Body)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("Saved audio file %s.m4a\n", audioUUID)
	fmt.Fprintf(w, "M4A uploaded successfully to %s", filepath)
}

func apiClientIos(w http.ResponseWriter, r *http.Request, u *shared.User) {
	query := database.NewQuery()
	query.Limit = 0
	sources, err := database.MosquitoSourceQuery(&query)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	requests, err := database.ServiceRequestQuery(&query)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	traps, err := database.TrapDataQuery(&query)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}

	response := NewResponseClientIos(sources, requests, traps)
	if err := render.Render(w, r, response); err != nil {
		render.Render(w, r, errRender(err))
		return
	}
}

func apiClientIosNotePut(w http.ResponseWriter, r *http.Request, u *shared.User) {
	id := chi.URLParam(r, "uuid")
	noteUUID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Failed to decode the uuid", http.StatusBadRequest)
		return
	}
	var payload shared.NidusNotePayload
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read the payload", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Println("Note PUT JSON decode error: ", err)
		output, err := os.OpenFile("/tmp/request.body", os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Println("Failed to open temp request.bady")
		}
		defer output.Close()
		output.Write(body)
		log.Println("Wrote request to /tmp/request.body")

		http.Error(w, "Failed to decode the payload", http.StatusBadRequest)
		return
	}
	if err := database.NoteUpdate(context.Background(), noteUUID, payload); err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func apiImagePost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	id := chi.URLParam(r, "uuid")
	noteUUID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Failed to decode the uuid", http.StatusBadRequest)
		return
	}

	var payload shared.NoteImagePayload
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read the payload", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Println("Image note POST JSON decode error: ", err)
		output, err := os.OpenFile("/tmp/request.body", os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Println("Failed to open temp request.bady")
		}
		defer output.Close()
		output.Write(body)
		log.Println("Wrote request to /tmp/request.body")

		http.Error(w, "Failed to decode the payload", http.StatusBadRequest)
		return
	}
	err = database.NoteImageCreate(context.Background(), noteUUID, payload, u.ID)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func apiImageContentPost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	u_str := chi.URLParam(r, "uuid")
	imageUUID, err := uuid.Parse(u_str)
	if err != nil {
		log.Println("Failed to parse image UUID", u_str)
		http.Error(w, "Failed to parse image UUID", http.StatusBadRequest)
	}
	// Read first 8 bytes to check PNG signature
	config, err := fssync.ReadConfig()
	if err != nil {
		log.Printf("Failed to read config", err)
		http.Error(w, "Unable to create file", http.StatusInternalServerError)
		return
	}
	filepath := fmt.Sprintf("%s/%s.photo", config.UserFiles.Directory, imageUUID.String())

	// Create file in configured directory
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Unable to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy rest of request body to file
	_, err = io.Copy(dst, r.Body)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("Saved image file %s\n", imageUUID)
	fmt.Fprintf(w, "PNG uploaded successfully to %s", filepath)
}

// / Test
func apiMosquitoSource(w http.ResponseWriter, r *http.Request, u *shared.User) {
	bounds, err := parseBounds(r)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}

	query := database.NewQuery()
	query.Bounds = *bounds
	query.Limit = 100
	sources, err := database.MosquitoSourceQuery(&query)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}

	data := []render.Renderer{}
	for _, s := range sources {
		data = append(data, NewResponseMosquitoSource(s))
	}
	if err := render.RenderList(w, r, data); err != nil {
		render.Render(w, r, errRender(err))
	}
}

func apiServiceRequest(w http.ResponseWriter, r *http.Request, u *shared.User) {
	bounds, err := parseBounds(r)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	query := database.NewQuery()
	query.Bounds = *bounds
	query.Limit = 100
	requests, err := database.ServiceRequestQuery(&query)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}

	data := []render.Renderer{}
	for _, sr := range requests {
		data = append(data, NewResponseServiceRequest(sr))
	}
	if err := render.RenderList(w, r, data); err != nil {
		render.Render(w, r, errRender(err))
	}
}

func apiTrapData(w http.ResponseWriter, r *http.Request, u *shared.User) {
	bounds, err := parseBounds(r)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}

	query := database.NewQuery()
	query.Bounds = *bounds
	query.Limit = 100
	trap_data, err := database.TrapDataQuery(&query)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}

	data := []render.Renderer{}
	for _, td := range trap_data {
		data = append(data, NewResponseTrapDatum(td))
	}
	if err := render.RenderList(w, r, data); err != nil {
		render.Render(w, r, errRender(err))
	}
}

func index(w http.ResponseWriter, r *http.Request, u *shared.User) {
	count, err := database.ServiceRequestCount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := html.PageDataIndex{
		ServiceRequestCount: count,
		Title:               "Gateway Sync Test",
		User:                u,
	}

	err = html.Index(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loginGet(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	fmt.Println("urlparam next:", next)
	err := html.Login(w, next)
	if err != nil {
		render.Render(w, r, errRender(err))
	}
}

func loginPost(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	if username == "" || password == "" {
		if username == "" {
			http.Error(w, "Missing username", http.StatusBadRequest)
		}
		if password == "" {
			http.Error(w, "Missing password", http.StatusBadRequest)
		}
		return
	}
	user, err := database.ValidateUser(username, password)
	if err != nil {
		w.Header().Set("WWW-Authenticate-Error", "invalid-credentials")
		http.Error(w, "Invalid username/password pair", http.StatusUnauthorized)
		return
	} else if user == nil {
		w.Header().Set("WWW-Authenticate-Error", "invalid-credentials")
		log.Println("Login for", username, "is invalid")
		http.Error(w, "Invalid username/password pair", http.StatusUnauthorized)
	}

	fmt.Println("Setting user session via login", user.DisplayName, user.ID, username)
	sessionManager.Put(r.Context(), "display_name", user.DisplayName)
	sessionManager.Put(r.Context(), "user_id", user.ID)
	sessionManager.Put(r.Context(), "username", username)
	if next == "" {
		w.WriteHeader(202)
	} else {
		http.Redirect(w, r, "/" + next, http.StatusFound)
	}
}

func logoutGet(w http.ResponseWriter, r *http.Request) {
	sessionManager.Put(r.Context(), "display_name", "")
	sessionManager.Put(r.Context(), "username", "")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func processAudioGet(w http.ResponseWriter, r *http.Request, u *shared.User) {
	query := database.NewQuery()
	query.Limit = 0
	audioNotes, err := database.NoteAudioQuery(&query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := html.PageDataProcessAudio{
		AudioNotes:  audioNotes,
		User:                u,
	}

	err = html.ProcessAudio(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

