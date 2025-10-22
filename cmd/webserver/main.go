package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/Gleipnir-Technology/fieldseeker-sync"
	"github.com/Gleipnir-Technology/fieldseeker-sync/database"
	"github.com/Gleipnir-Technology/fieldseeker-sync/html"
	"github.com/Gleipnir-Technology/fieldseeker-sync/shared"
)

var sessionManager *scs.SessionManager

//go:embed static
var embeddedStaticFS embed.FS

func errRender(err error) render.Renderer {
	fmt.Println("Rendering error:", err)
	return &ResponseErr{
		Error:          err,
		HTTPStatusCode: 500,
		StatusText:     "Error rendering response",
		ErrorText:      err.Error(),
	}
}

func main() {
	log.Fatal(run())
}

func run() error {
	err := sentry.Init(sentry.ClientOptions{
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		return err
	}
	defer sentry.Flush(2 * time.Second)

	err = fssync.InitDB()
	if err != nil {
		fmt.Printf("Failed to init database: %v", err)
		os.Exit(1)
	}

	fssync.StartAudioWorker(context.Background())
	err = fssync.StartLabelStudioWorker(context.Background())
	if err != nil {
		fmt.Printf("Failed to create label studio processor: %v", err)
		os.Exit(2)
	}

	sessionManager = scs.New()
	sessionManager.Store = pgxstore.New(database.PGInstance.DB)
	sessionManager.Lifetime = 24 * time.Hour

	// Set our own responder so that we can set headers ourselves
	render.Respond = Responder
	sentryMiddleware := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// Sentry goes after recoverer with repanic
	r.Use(sentryMiddleware.Handle)
	r.Use(sessionManager.LoadAndSave)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	//html.InitializeTemplates()
	r.Method("GET", "/", NewEnsureAuth(index))
	r.Method("GET", "/audio/{uuid}.{extension}", NewEnsureAuth(audioGet))
	r.Method("GET", "/process-audio", NewEnsureAuth(processAudioGet))
	r.Method("GET", "/process-audio/{id}", NewEnsureAuth(processAudioIdGet))
	r.Method("POST", "/process-audio/{id}", NewEnsureAuth(processAudioIdPost))
	r.Method("POST", "/process-audio/{id}/delete", NewEnsureAuth(processAudioIdDeletePost))
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
	localFS := http.Dir("./static")
	FileServer(r, "/static", localFS, embeddedStaticFS, "static")

	bind := os.Getenv("FIELDSEEKER_SYNC_WEBSERVER_BIND")
	if len(bind) == 0 {
		bind = ":3000"
	}
	log.Println("Serving web requests on", bind)
	return http.ListenAndServe(bind, r)
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
	sr := html.ContentServiceRequests{
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
func FileServer(r chi.Router, path string, root http.FileSystem, embeddedFS embed.FS, embeddedPath string) {
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

		// Determine the actual file path
		requestedPath := strings.TrimPrefix(r.URL.Path, pathPrefix)

		// Try to open from local filesystem first for development
		localFile, localErr := root.Open("cmd/webserver/static" + requestedPath)

		var fileToServe http.File

		if localErr == nil {
			// File found in local filesystem
			fileToServe = localFile
		} else {
			// If not found locall, try embedded filesystem
			embeddedFilePath := filepath.Join(embeddedPath, requestedPath)
			embeddedFile, err := embeddedFS.Open(embeddedFilePath)

			if err != nil {
				http.NotFound(w, r)
				return
			}

			// Wrap the embedded file to implement http.File interface
			fileToServe = &embeddedFileWrapper{embeddedFile}

		}

		// Create a custom ResponseWriter that allows us to modify headers
		crw := &customResponseWriter{ResponseWriter: w}

		// Serve the file
		http.ServeContent(crw, r, requestedPath, time.Time{}, fileToServe)

		// Close the file
		fileToServe.Close()
	})
}

// Custom ResponseWriter to track Content-Type
type customResponseWriter struct {
	http.ResponseWriter
	contentType string
	wroteHeader bool
}

func (crw *customResponseWriter) WriteHeader(code int) {
	crw.wroteHeader = true
	crw.ResponseWriter.WriteHeader(code)
}

func (crw *customResponseWriter) Header() http.Header {
	return crw.ResponseWriter.Header()
}

func (crw *customResponseWriter) Write(b []byte) (int, error) {
	if !crw.wroteHeader {
		if crw.contentType == "" {
			crw.contentType = http.DetectContentType(b)
			crw.ResponseWriter.Header().Set("Content-Type", crw.contentType)
		}
		crw.WriteHeader(http.StatusOK)
	}
	return crw.ResponseWriter.Write(b)
}

type embeddedFileWrapper struct {
	file fs.File
}

func (e *embeddedFileWrapper) Close() error {
	return e.file.Close()
}

func (e *embeddedFileWrapper) Read(p []byte) (n int, err error) {
	return e.file.Read(p)
}

type Seeker interface {
	Seek(offset int64, whence int) (int64, error)
}

func (e *embeddedFileWrapper) Seek(offset int64, whence int) (int64, error) {
	if seeker, ok := e.file.(Seeker); ok {
		return seeker.Seek(offset, whence)
	}
	return 0, fmt.Errorf("Seek not supported")
}

func (e *embeddedFileWrapper) Readdir(count int) ([]os.FileInfo, error) {
	// This is a bit tricky with embedded files
	if dirFile, ok := e.file.(fs.ReadDirFile); ok {
		entries, err := dirFile.ReadDir(count)
		if err != nil {
			return nil, err
		}

		fileInfos := make([]os.FileInfo, len(entries))
		for i, entry := range entries {
			fileInfos[i], err = entry.Info()
			if err != nil {
				return nil, err
			}
		}
		return fileInfos, nil
	}
	return nil, fmt.Errorf("Readdir not supported")
}

func (e *embeddedFileWrapper) Stat() (os.FileInfo, error) {
	return e.file.Stat()
}
