package main

import (
	"log"
	"net/http"

	"github.com/go-chi/render"

	"gleipnir.technology/fieldseeker-sync"
	"gleipnir.technology/fieldseeker-sync/html"
)

func apiClientIos(w http.ResponseWriter, r *http.Request, u *fssync.User) {
	query := fssync.NewQuery()
	query.Limit = 0
	sources, err := fssync.MosquitoSourceQuery(&query)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	requests, err := fssync.ServiceRequestQuery(&query)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	traps, err := fssync.TrapDataQuery(&query)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}

	response := NewResponseClientIos(sources, requests, traps)
	if err := render.Render(w, r, response); err != nil {
		render.Render(w, r, errRender(err))
	}
}

func apiMosquitoSource(w http.ResponseWriter, r *http.Request, u *fssync.User) {
	bounds, err := parseBounds(r)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}

	query := fssync.NewQuery()
	query.Bounds = *bounds
	query.Limit = 100
	sources, err := fssync.MosquitoSourceQuery(&query)
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

func apiServiceRequest(w http.ResponseWriter, r *http.Request, u *fssync.User) {
	bounds, err := parseBounds(r)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	query := fssync.NewQuery()
	query.Bounds = *bounds
	query.Limit = 100
	requests, err := fssync.ServiceRequestQuery(&query)
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

func apiTrapData(w http.ResponseWriter, r *http.Request, u *fssync.User) {
	bounds, err := parseBounds(r)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}

	query := fssync.NewQuery()
	query.Bounds = *bounds
	query.Limit = 100
	trap_data, err := fssync.TrapDataQuery(&query)
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

func index(w http.ResponseWriter, r *http.Request, u *fssync.User) {
	count, err := fssync.ServiceRequestCount()
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
	user, err := fssync.ValidateUser(username, password)
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
