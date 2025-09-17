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
	//"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	"github.com/Gleipnir-Technology/fieldseeker-sync"
	"github.com/Gleipnir-Technology/fieldseeker-sync/database"
	"github.com/Gleipnir-Technology/fieldseeker-sync/database/models"
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

	config := fssync.ReadConfig()
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
	config := fssync.ReadConfig()
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

func audioGet(w http.ResponseWriter, r *http.Request, u *shared.User) {
	extension := chi.URLParam(r, "extension")
	uuid := chi.URLParam(r, "uuid")
	config := fssync.ReadConfig()
	filePath := "unknown"
	contentType := "unknown"
	if extension == "m4a" {
		contentType = "audio/mpeg"
		filePath = fmt.Sprintf("%s/%s-normalized.m4a", config.UserFiles.Directory, uuid)
	} else if extension == "mp3" {
		contentType = "audio/mp3"
		filePath = fmt.Sprintf("%s/%s.mp3", config.UserFiles.Directory, uuid)
	} else if extension == "ogg" {
		contentType = "audio/ogg"
		filePath = fmt.Sprintf("%s/%s.ogg", config.UserFiles.Directory, uuid)
	} else {
		http.Error(w, fmt.Sprintf("Extension '%s' not found", extension), http.StatusNotFound)
		return
	}
	log.Printf("Serving %s", filePath)
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Audio file not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fileSize := fileInfo.Size()

	// Parse range header
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		ranges, err := parseRange(rangeHeader, fileSize)
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestedRangeNotSatisfiable)
			return
		}
		// We'll handle just the first range for this example
		if len(ranges) > 0 {
			start, end := ranges[0].start, ranges[0].end

			// Seek to the start position
			if _, err := file.Seek(start, io.SeekStart); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Set headers for partial content
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", end-start+1))
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Accept-Ranges", "bytes")
			w.WriteHeader(http.StatusPartialContent)

			// Send the partial content
			if _, err := io.CopyN(w, file, end-start+1); err != nil {
				log.Printf("Error streaming partial content: %v", err)
				return
			}
			return
		}
	}

	// If no range, serve the whole file
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")

	// Copy the file to the response writer
	if _, err := io.Copy(w, file); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func index(w http.ResponseWriter, r *http.Request, u *shared.User) {
	count, err := database.ServiceRequestCount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := html.ContentIndex{
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
		http.Redirect(w, r, next, http.StatusFound)
	}
}

func logoutGet(w http.ResponseWriter, r *http.Request) {
	sessionManager.Put(r.Context(), "display_name", "")
	sessionManager.Put(r.Context(), "username", "")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

type httpRange struct {
	start, end int64
}

func parseRange(rangeHeader string, fileSize int64) ([]httpRange, error) {
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return nil, fmt.Errorf("invalid range header format")
	}

	rangeHeader = strings.TrimPrefix(rangeHeader, "bytes=")
	ranges := strings.Split(rangeHeader, ",")
	parsedRanges := make([]httpRange, 0, len(ranges))

	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}

		parts := strings.Split(r, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range format")
		}

		var start, end int64
		var err error

		if parts[0] == "" {
			// suffix-length format: -N
			end = fileSize - 1
			if start, err = strconv.ParseInt(parts[1], 10, 64); err != nil {
				return nil, fmt.Errorf("invalid range start")
			}
			start = fileSize - start
			if start < 0 {
				start = 0
			}
		} else {
			// standard format: N-M
			if start, err = strconv.ParseInt(parts[0], 10, 64); err != nil {
				return nil, fmt.Errorf("invalid range start")
			}
			if parts[1] == "" {
				// N- format
				end = fileSize - 1
			} else {
				// N-M format
				if end, err = strconv.ParseInt(parts[1], 10, 64); err != nil {
					return nil, fmt.Errorf("invalid range end")
				}
			}
		}

		if start > end || start >= fileSize {
			return nil, fmt.Errorf("invalid range: start after end or file size")
		}
		if end >= fileSize {
			end = fileSize - 1
		}

		parsedRanges = append(parsedRanges, httpRange{start, end})
	}

	return parsedRanges, nil
}

func processAudioGet(w http.ResponseWriter, r *http.Request, u *shared.User) {
	sortField := r.URL.Query().Get("sort")
	var sortEnum database.TaskAudioReviewOutstandingSort
	switch sortField {
	case "":
		sortEnum = database.SortCreated
	case "created_at":
		sortEnum = database.SortCreated
	case "duration":
		sortEnum = database.SortAudioDuration
	case "creator":
		sortEnum = database.SortCreatorName
	case "needs_review":
		sortEnum = database.SortNeedsReview
	default:
		sortEnum = database.SortCreated
	}
	sortOrder := r.URL.Query().Get("order")
	if sortOrder == "" {
		sortOrder = "asc"
	}

	rows, err := database.TaskAudioReviewList(sortEnum, sortOrder == "asc")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//sort.Sort(byReviewedAndAge(tasks))
	usersById, err := usersById()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := html.ContentProcessAudio{
		Rows:      rows,
		SortField: sortField,
		SortOrder: sortOrder,
		UsersById: usersById,
		User:      u,
	}

	err = html.ProcessAudio(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func processAudioIdGet(w http.ResponseWriter, r *http.Request, u *shared.User) {
	id_str := chi.URLParam(r, "id")
	id, err := strconv.Atoi(id_str)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task, err := models.FindTaskAudioReview(context.Background(), database.PGInstance.BobDB, int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	noteAudio, err := database.NoteAudioGetLatest(context.Background(), task.NoteAudioUUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	usersById, err := usersById()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := html.ContentProcessAudioId{
		NoteAudio: noteAudio,
		Task:      task,
		UsersById: usersById,
		User:      u,
	}

	log.Printf("noteAudio %s isvalue %v", noteAudio.UUID, noteAudio.Transcription.IsNull())
	err = html.ProcessAudioId(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func processAudioIdPost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	uuid := chi.URLParam(r, "uuid")

	r.ParseForm()
	transcription := r.Form.Get("transcription")
	log.Printf("Updating %s to transcript %s", uuid, transcription)
	err := database.NoteAudioUpdateTranscription(uuid, transcription, u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/process-audio/"+uuid, http.StatusFound)
}

func processAudioIdDeletePost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	uuid := chi.URLParam(r, "uuid")
	log.Printf("Deleting %s", uuid)
	err := database.NoteAudioUpdateDelete(uuid, u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/process-audio", http.StatusFound)
}

func processAudioIdReviewedPost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	uuid := chi.URLParam(r, "uuid")
	log.Printf("Updating %s to reviewed without changes", uuid)
	err := database.NoteAudioUpdateReviewed(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/process-audio", http.StatusFound)
}

func processAudioIdNeedsFurtherReviewPost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	uuid := chi.URLParam(r, "uuid")
	log.Printf("Updating %s to needs further review", uuid)
	err := database.NoteAudioUpdateNeedsFurtherReview(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/process-audio", http.StatusFound)
}

func usersById() (map[int]*shared.User, error) {
	users, err := database.Users()
	if err != nil {
		return nil, err
	}
	usersById := make(map[int]*shared.User)
	for _, u := range users {
		usersById[u.ID] = u
	}
	return usersById, nil
}

type byReviewedAndAge []*models.TaskAudioReview

func (a byReviewedAndAge) Len() int      { return len(a) }
func (a byReviewedAndAge) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byReviewedAndAge) Less(i, j int) bool {
	if a[i].NeedsReview == a[j].NeedsReview {
		if a[i].ReviewedBy == a[j].ReviewedBy {
			return a[i].Created.Before(a[j].Created)
		}
		return a[i].ReviewedBy.IsNull()
	}
	return a[i].NeedsReview
}
