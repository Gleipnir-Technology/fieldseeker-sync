package main

import (
	"fmt"
	"net/http"

	"github.com/Gleipnir-Technology/fieldseeker-sync/database"
	"github.com/Gleipnir-Technology/fieldseeker-sync/shared"
)

type NoCredentialsError struct{}

func (e NoCredentialsError) Error() string { return "No credentials were present in the request" }

type InvalidSessionError struct{}

func (e InvalidSessionError) Error() string {
	return "A session was present, but the contents were not valid"
}

// authenticatedHandler is a handler function that also requires a user
type AuthenticatedHandler func(http.ResponseWriter, *http.Request, *shared.User)

type EnsureAuth struct {
	handler AuthenticatedHandler
}

func NewEnsureAuth(handlerToWrap AuthenticatedHandler) *EnsureAuth {
	return &EnsureAuth{handlerToWrap}
}

func (ea *EnsureAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If this is an API request respond with a more machine-readable error state
	accept := r.Header.Values("Accept")
	offers := []string{"application/json", "text/html"}

	content_type := NegotiateContent(accept, offers)
	user, err := getAuthenticatedUser(r)
	if err != nil {
		var msg []byte
		// Separate return codes for different authentication failures
		if _, ok := err.(*NoCredentialsError); ok {
			fmt.Println("No credentials present and no session")
			w.Header().Set("WWW-Authenticate-Error", "no-credentials")
			msg = []byte("Please provide credentials.\n")
		} else if _, ok := err.(*database.NoUserError); ok {
			w.Header().Set("WWW-Authenticate-Error", "invalid-credentials")
			msg = []byte("Invalid credentials provided.\n")
		} else if _, ok := err.(*database.PasswordVerificationError); ok {
			w.Header().Set("WWW-Authenticate-Error", "invalid-credentials")
			msg = []byte("Invalid credentials provided.\n")
		} else if _, ok := err.(*InvalidSessionError); ok {
			fmt.Println("Got an invalid session, this usually indicates a particularly bad connection or application logic error")
		}

		if content_type == "text/html" {
			http.Redirect(w, r, "/login?next="+r.URL.Path, http.StatusSeeOther)
			return
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="Nidus Sync"`)
		w.WriteHeader(401)
		w.Write(msg)
		return
	}

	ea.handler(w, r, user)
}

func getAuthenticatedUser(r *http.Request) (*shared.User, error) {
	// See if we can get the user from the session first
	display_name := sessionManager.GetString(r.Context(), "display_name")
	user_id := sessionManager.GetInt(r.Context(), "user_id")
	username := sessionManager.GetString(r.Context(), "username")
	fmt.Printf("Session data '%s' %d '%s'\n", display_name, user_id, username)
	if user_id > 0 && len(display_name) > 0 && len(username) > 0 {
		return &shared.User{
			DisplayName: display_name,
			ID:          user_id,
			Username:    username,
		}, nil
	}

	// If we can't get it from the session, let's see if we can authenticate from
	// the header
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, &NoCredentialsError{}
	}
	user, err := database.ValidateUser(username, password)
	if err != nil {
		return nil, err
	}

	fmt.Println("Setting user session", user.DisplayName, user.ID, username)
	sessionManager.Put(r.Context(), "display_name", user.DisplayName)
	sessionManager.Put(r.Context(), "user_id", user.ID)
	sessionManager.Put(r.Context(), "username", username)

	return user, nil
}
