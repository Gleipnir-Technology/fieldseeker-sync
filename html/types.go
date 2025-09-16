package html

import (
	"github.com/Gleipnir-Technology/fieldseeker-sync/database/models"
	"github.com/Gleipnir-Technology/fieldseeker-sync/shared"
)

type ContentIndex struct {
	ServiceRequestCount int
	Title               string
	User                *shared.User
}

type ContentLogin struct {
	Next  string
	Title string
	User  *shared.User
}

type ContentProcessAudio struct {
	Tasks     []*models.TaskAudioReview
	UsersById map[int]*shared.User
	User      *shared.User
}

type ContentProcessAudioId struct {
	AudioNote *shared.NoteAudio
	UsersById map[int]*shared.User
	User      *shared.User
}

type ContentServiceRequests struct {
	ServiceRequests []shared.ServiceRequest
	User            *shared.User
}
