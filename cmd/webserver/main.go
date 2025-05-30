package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"gleipnir.technology/fieldseeker-sync"
	"gleipnir.technology/fieldseeker-sync/html"
)

// authenticatedHandler is a handler function that also requires a user
type AuthenticatedHandler func(http.ResponseWriter, *http.Request, *fssync.User)

type EnsureAuth struct {
	handler AuthenticatedHandler
}

func (ea *EnsureAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := getAuthenticatedUser(r)
	if err != nil {
		http.Redirect(w, r, "/login?next="+r.URL.Path, http.StatusSeeOther)
		return
	}

	ea.handler(w, r, user)
}

func NewEnsureAuth(handlerToWrap AuthenticatedHandler) *EnsureAuth {
	return &EnsureAuth{handlerToWrap}
}

var sessionManager *scs.SessionManager

func errRender(err error) render.Renderer {
	return &ResponseErr{
		Error:          err,
		HTTPStatusCode: 500,
		StatusText:     "Error rendering response",
		ErrorText:      err.Error(),
	}
}

func getAuthenticatedUser(r *http.Request) (*fssync.User, error) {
	display_name := sessionManager.GetString(r.Context(), "display_name")
	username := sessionManager.GetString(r.Context(), "username")
	if display_name == "" || username == "" {
		return nil, errors.New("No valid user in session")
	}
	return &fssync.User{
		DisplayName: display_name,
		Username:    username,
	}, nil
}

func main() {
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(sessionManager.LoadAndSave)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	err := fssync.InitDB()
	if err != nil {
		fmt.Println("Failed to init fssync: %v", err)
	}

	//html.InitializeTemplates()
	r.Method("GET", "/", NewEnsureAuth(index))
	r.Method("GET", "/service-request", NewEnsureAuth(serviceRequestList))

	r.Get("/login", loginGet)
	r.Post("/login", loginPost)
	r.Get("/logout", logoutGet)

	r.Route("/api", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))
		r.Method("GET", "/mosquito-source", NewEnsureAuth(mosquitoSourceApi))
		r.Method("GET", "/service-request", NewEnsureAuth(serviceRequestApi))
		r.Method("GET", "/trap-data", NewEnsureAuth(trapDataApi))
		r.Method("GET", "/client/ios", NewEnsureAuth(clientIosApi))
		r.Get("/webhook/fieldseeker", webhookFieldseeker)
		r.Post("/webhook/fieldseeker", webhookFieldseeker)
	})
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	FileServer(r, "/static", filesDir)

	log.Println("Serving web requests on :3000")
	http.ListenAndServe(":3000", r)
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

func mosquitoSourceApi(w http.ResponseWriter, r *http.Request, u *fssync.User) {
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

func parseBounds(r *http.Request) (*fssync.Bounds, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	east := r.FormValue("east")
	north := r.FormValue("north")
	south := r.FormValue("south")
	west := r.FormValue("west")

	bounds := fssync.Bounds{}

	var temp float64
	temp, err = strconv.ParseFloat(east, 64)
	if err != nil {
		return nil, err
	}
	bounds.East = temp
	temp, err = strconv.ParseFloat(north, 64)
	if err != nil {
		return nil, err
	}
	bounds.North = temp
	temp, err = strconv.ParseFloat(south, 64)
	if err != nil {
		return nil, err
	}
	bounds.South = temp
	temp, err = strconv.ParseFloat(west, 64)
	if err != nil {
		return nil, err
	}
	bounds.West = temp
	return &bounds, nil
}

func serviceRequestApi(w http.ResponseWriter, r *http.Request, u *fssync.User) {
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

func clientIosApi(w http.ResponseWriter, r *http.Request, u *fssync.User) {
	notes, err := fssync.NoteQuery()
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	data := []render.Renderer{}
	for _, n := range notes {
		data = append(data, NewNote(n))
	}
	if err := render.RenderList(w, r, data); err != nil {
		render.Render(w, r, errRender(err))
	}
}
func serviceRequestList(w http.ResponseWriter, r *http.Request, u *fssync.User) {
	bounds := fssync.Bounds{
		East:  -180,
		North: 180,
		South: -180,
		West:  180,
	}
	query := fssync.NewQuery()
	query.Bounds = bounds
	query.Limit = 100
	requests, err := fssync.ServiceRequestQuery(&query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sr := html.PageDataServiceRequests{
		ServiceRequests: requests,
		User:            u,
	}
	err = html.ServiceRequests(w, sr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func trapDataApi(w http.ResponseWriter, r *http.Request, u *fssync.User) {
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

func webhookFieldseeker(w http.ResponseWriter, r *http.Request) {
	// Create or open the log file
	file, err := os.OpenFile("webhook/request.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Write timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(file, "\n=== Request logged at %s ===\n", timestamp)

	// Write request line
	fmt.Fprintf(file, "%s %s %s\n", r.Method, r.RequestURI, r.Proto)

	// Write all headers
	fmt.Fprintf(file, "\nHeaders:\n")
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Fprintf(file, "%s: %s\n", name, value)
		}
	}

	// Write body
	fmt.Fprintf(file, "\nBody:\n")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		fmt.Fprintf(file, "Error reading body: %v\n", err)
	} else {
		file.Write(body)
		if len(body) == 0 {
			fmt.Fprintf(file, "(empty body)")
		}
	}

	fmt.Fprintf(file, "\n=== End of request ===\n\n")

	// Extract the crc_token value for the signature portion

	// Respond with 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
