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

	"github.com/Gleipnir-Technology/fieldseeker-sync"
	"github.com/Gleipnir-Technology/fieldseeker-sync/database"
	"github.com/Gleipnir-Technology/fieldseeker-sync/shared"
	"github.com/Gleipnir-Technology/fieldseeker-sync/html"
)

// authenticatedHandler is a handler function that also requires a user
type AuthenticatedHandler func(http.ResponseWriter, *http.Request, *shared.User)

type EnsureAuth struct {
	handler AuthenticatedHandler
}

func NewEnsureAuth(handlerToWrap AuthenticatedHandler) *EnsureAuth {
	return &EnsureAuth{handlerToWrap}
}

var sessionManager *scs.SessionManager

func (ea *EnsureAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If this is an API request respond with a more machine-readable error state
	accept := r.Header.Values("Accept")
	offers := []string{"application/json", "text/html"}

	content_type := NegotiateContent(accept, offers)
	user, err := getAuthenticatedUser(r)
	if err != nil {
		if content_type == "text/html" {
			http.Redirect(w, r, "/login?next="+r.URL.Path, http.StatusSeeOther)
			return
		} else {
			http.Error(w, "Login required", http.StatusUnauthorized)
			return
		}
	}

	ea.handler(w, r, user)
}

func errRender(err error) render.Renderer {
	return &ResponseErr{
		Error:          err,
		HTTPStatusCode: 500,
		StatusText:     "Error rendering response",
		ErrorText:      err.Error(),
	}
}

func getAuthenticatedUser(r *http.Request) (*shared.User, error) {
	display_name := sessionManager.GetString(r.Context(), "display_name")
	username := sessionManager.GetString(r.Context(), "username")
	if display_name == "" || username == "" {
		return nil, errors.New("No valid user in session")
	}
	return &shared.User{
		DisplayName: display_name,
		Username:    username,
	}, nil
}

func main() {
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	// Set our own responder so that we can set headers ourselves
	render.Respond = Responder
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
		fmt.Printf("Failed to init shared: %v", err)
		os.Exit(1)
	}

	//html.InitializeTemplates()
	r.Method("GET", "/", NewEnsureAuth(index))
	r.Method("GET", "/service-request", NewEnsureAuth(serviceRequestList))

	r.Get("/login", loginGet)
	r.Post("/login", loginPost)
	r.Get("/logout", logoutGet)

	r.Route("/api", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))
		r.Method("GET", "/mosquito-source", NewEnsureAuth(apiMosquitoSource))
		r.Method("GET", "/service-request", NewEnsureAuth(apiServiceRequest))
		r.Method("GET", "/trap-data", NewEnsureAuth(apiTrapData))
		r.Method("GET", "/client/ios", NewEnsureAuth(apiClientIos))
		r.Method("PUT", "/client/ios/note/{uuid}", NewEnsureAuth(apiClientIosNotePut))
		r.Method("POST", "/audio/{uuid}", NewEnsureAuth(apiAudioPost))
		r.Method("POST", "/audio/{uuid}/content", NewEnsureAuth(apiAudioContentPost))
		r.Method("POST", "/image/{uuid}", NewEnsureAuth(apiImagePost))
		r.Method("POST", "/image/{uuid}/content", NewEnsureAuth(apiImageContentPost))
		r.Get("/webhook/fieldseeker", webhookFieldseeker)
		r.Post("/webhook/fieldseeker", webhookFieldseeker)
	})
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	FileServer(r, "/static", filesDir)

	bind := os.Getenv("FIELDSEEKER_SYNC_WEBSERVER_BIND")
	if len(bind) == 0 {
		bind = ":3000"
	}
	log.Println("Serving web requests on", bind)
	http.ListenAndServe(bind, r)
}

func parseBounds(r *http.Request) (*shared.Bounds, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	east := r.FormValue("east")
	north := r.FormValue("north")
	south := r.FormValue("south")
	west := r.FormValue("west")

	bounds := shared.Bounds{}

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

func serviceRequestList(w http.ResponseWriter, r *http.Request, u *shared.User) {
	bounds := shared.Bounds{
		East:  -180,
		North: 180,
		South: -180,
		West:  180,
	}
	query := database.NewQuery()
	query.Bounds = bounds
	query.Limit = 100
	requests, err := database.ServiceRequestQuery(&query)
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
