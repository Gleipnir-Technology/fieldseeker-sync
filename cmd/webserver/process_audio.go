package main

import (
	"context"
	"log"
	"net/http"
	//"sort"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Gleipnir-Technology/fieldseeker-sync/database"
	"github.com/Gleipnir-Technology/fieldseeker-sync/database/models"
	"github.com/Gleipnir-Technology/fieldseeker-sync/html"
	"github.com/Gleipnir-Technology/fieldseeker-sync/shared"
)

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

func idToAudioReviewTask(w http.ResponseWriter, r *http.Request) *models.TaskAudioReview {
	id_str := chi.URLParam(r, "id")
	id, err := strconv.Atoi(id_str)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	task, err := models.FindTaskAudioReview(context.Background(), database.PGInstance.BobDB, int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return task
}

func processAudioIdGet(w http.ResponseWriter, r *http.Request, u *shared.User) {
	task := idToAudioReviewTask(w, r)
	if task == nil {
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
	task := idToAudioReviewTask(w, r)
	if task == nil {
		return
	}

	r.ParseForm()
	transcription := r.Form.Get("transcription")
	log.Printf("Updating %s to transcript %s", task.NoteAudioUUID, transcription)
	err := database.NoteAudioUpdateTranscription(task.NoteAudioUUID, transcription, u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/process-audio", http.StatusFound)
}

func processAudioIdDeletePost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	task := idToAudioReviewTask(w, r)
	if task == nil {
		return
	}
	log.Printf("Deleting %s", task.NoteAudioUUID)
	err := database.NoteAudioUpdateDelete(task.NoteAudioUUID, u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/process-audio", http.StatusFound)
}

func processAudioIdReviewedPost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	task := idToAudioReviewTask(w, r)
	if task == nil {
		return
	}
	err := database.NoteAudioUpdateReviewed(task.NoteAudioUUID, u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/process-audio", http.StatusFound)
}

func processAudioIdFurtherReviewedPost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	task := idToAudioReviewTask(w, r)
	if task == nil {
		return
	}
	log.Printf("Updating %s to further reviewed", task.NoteAudioUUID)
	err := database.NoteAudioUpdateFurtherReviewed(task.ID, u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/process-audio", http.StatusFound)
}

func processAudioIdNeedsFurtherReviewPost(w http.ResponseWriter, r *http.Request, u *shared.User) {
	task := idToAudioReviewTask(w, r)
	if task == nil {
		return
	}
	log.Printf("Updating %s to needs further review", task.NoteAudioUUID)
	err := database.NoteAudioUpdateNeedsFurtherReview(task.NoteAudioUUID, u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/process-audio", http.StatusFound)
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
