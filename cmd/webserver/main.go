package main

import (
	"errors"
	"fmt"
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

	"gleipnir.technology/fieldseeker-sync-bridge"
	"gleipnir.technology/fieldseeker-sync-bridge/html"
)

// ErrResponse renderer type for handling all sorts of errors.
//
// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ErrResponse struct {
	Error          error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

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

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func errRender(err error) render.Renderer {
	return &ErrResponse{
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
		r.Method("GET", "/service-request", NewEnsureAuth(serviceRequestApi))
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
	log.Println("Login attempt", username, password)
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

type ServiceRequestResponse struct {
	Address  *string `json:"address"`
	City     *string `json:"city"`
	Lat      float64 `json:"lat"`
	Long     float64 `json:"long"`
	Priority *string `json:"priority"`
	Source   *string `json:"source"`
	Status   *string `json:"status"`
	Target   *string `json:"target"`
	Zip      *string `json:"zip"`
}

func (srr ServiceRequestResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
func NewServiceRequest(sr *fssync.ServiceRequest) ServiceRequestResponse {
	return ServiceRequestResponse{
		Address:  sr.Address,
		City:     sr.City,
		Lat:      sr.Geometry.Y,
		Long:     sr.Geometry.X,
		Priority: sr.Priority,
		Status:   sr.Status,
		Source:   sr.Source,
		Target:   sr.Target,
		Zip:      sr.Zip,
	}
}
func serviceRequestApi(w http.ResponseWriter, r *http.Request, u *fssync.User) {
	err := r.ParseForm()
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}

	east := r.FormValue("east")
	north := r.FormValue("north")
	south := r.FormValue("south")
	west := r.FormValue("west")

	bounds := fssync.Bounds{}

	var temp float64
	temp, err = strconv.ParseFloat(east, 64)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	bounds.East = temp
	temp, err = strconv.ParseFloat(north, 64)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	bounds.North = temp
	temp, err = strconv.ParseFloat(south, 64)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	bounds.South = temp
	temp, err = strconv.ParseFloat(west, 64)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}
	bounds.West = temp

	requests, err := fssync.ServiceRequests(bounds)
	if err != nil {
		render.Render(w, r, errRender(err))
		return
	}

	data := []render.Renderer{}
	for _, sr := range requests {
		data = append(data, NewServiceRequest(sr))
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
	requests, err := fssync.ServiceRequests(bounds)
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
