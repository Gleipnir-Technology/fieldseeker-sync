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
	database.NoteAudioCreate(context.Background(), noteUUID, payload)
	w.WriteHeader(http.StatusAccepted)
}

func apiAudioContentPost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	u_str := chi.URLParam(r, "uuid")
	audioUUID, err := uuid.Parse(u_str)
	if err != nil {
		http.Error(w, "Failed to parse image UUID", http.StatusBadRequest)
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

   // Write the header bytes we already read
   _, err = io.Copy(dst, r.Body)
   if err != nil {
   	http.Error(w, "Unable to save file", http.StatusInternalServerError)
   	return
   }

   // Copy rest of request body to file
   _, err = io.Copy(dst, r.Body)
   if err != nil {
   	http.Error(w, "Unable to save file", http.StatusInternalServerError)
   	return
   }

   w.WriteHeader(http.StatusOK)
   log.Printf("Saved image file %s\n", audioUUID)
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
	database.NoteUpdate(context.Background(), noteUUID, payload)
	w.WriteHeader(http.StatusAccepted)
}

func apiImagePost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	u_str := chi.URLParam(r, "uuid")
	imageUUID, err := uuid.Parse(u_str)
	if err != nil {
		log.Println("Failed to parse image UUID", u_str)
		http.Error(w, "Failed to parse image UUID", http.StatusBadRequest)
	}
	// Read first 8 bytes to check PNG signature
	/*
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	buffer := make([]byte, 8)
	_, err = r.Body.Read(buffer)
	if err != nil {
		log.Println("Unable to read image request body", err)
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	// Verify PNG signature
	for i, b := range pngSignature {
		if buffer[i] != b {
			http.Error(w, "File is not a valid PNG", http.StatusBadRequest)
			return
		}
	}

	*/
	config, err := fssync.ReadConfig()
	if err != nil {
		log.Printf("Failed to read config", err)
		http.Error(w, "Unable to create file", http.StatusInternalServerError)
		return
	}
	filepath := fmt.Sprintf("%s/%s.png", config.UserFiles.Directory, imageUUID.String())

	// Create file in configured directory
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Unable to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Write the header bytes we already read
	_, err = io.Copy(dst, r.Body)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}

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
		Title:		     "Gateway Sync Test",
		User:		     u,
	}

	err = html.Index(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loginGet(w http.ResponseWriter, r *http.Request) {
	err := html.Login(w)
	if err != nil {
		render.Render(w, r, errRender(err))
	}
}

func loginPost(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Invalid username/password pair", http.StatusUnauthorized)
		return
	} else if user == nil {
		log.Println("Login for", username, "is invalid")
		http.Error(w, "Invalid username/password pair", http.StatusUnauthorized)
	}

	sessionManager.Put(r.Context(), "display_name", user.DisplayName)
	sessionManager.Put(r.Context(), "username", username)
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutGet(w http.ResponseWriter, r *http.Request) {
	sessionManager.Put(r.Context(), "display_name", "")
	sessionManager.Put(r.Context(), "username", "")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
