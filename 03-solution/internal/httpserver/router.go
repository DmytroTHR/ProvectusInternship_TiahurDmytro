package httpserver

import (
	"encoding/json"
	"net/http"
	"solution/internal/bucket"
	"solution/internal/model"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	Users        model.Users
	bucketClient *bucket.Client
}

func NewUserHandler(users model.Users, client *bucket.Client) *UserHandler {
	return &UserHandler{
		Users:        users,
		bucketClient: client,
	}
}

func NewRouter(userH *UserHandler) http.Handler {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		output, _ := json.MarshalIndent(&struct {
			Message string
		}{Message: "Wrong path selected"}, "", "\t")
		w.Write(output)
	})
	r.Get("/data", userH.getUsersData)
	r.Get("/stats", userH.getUsersStats)
	r.Post("/data", userH.postUsersData)
	return r
}

func (uh *UserHandler) postUsersData(w http.ResponseWriter, r *http.Request) {
	uh.bucketClient.UpdateCh <- struct{}{}
	output, _ := json.MarshalIndent(&struct {
		Message string
	}{Message: "Data is being updated, you can reGET it."}, "", "\t")
	w.Write(output)
}

func (uh *UserHandler) getUsersData(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	users, err := uh.filterUsers(w, r)
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusOK)
	output, _ := json.MarshalIndent(users, "", "\t")
	w.Write(output)
}

func (uh *UserHandler) getUsersStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	users, err := uh.filterUsers(w, r)
	if err != nil {
		return
	}
	w.WriteHeader(http.StatusOK)
	output, _ := json.MarshalIndent(users.CalculateAverage(), "", "\t")
	w.Write(output)
}

func (uh *UserHandler) filterUsers(w http.ResponseWriter, r *http.Request) (model.Users, error) {
	filters := getFilters(r)
	users, err := uh.bucketClient.FilterUsers(uh.Users, filters)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		output, _ := json.MarshalIndent(&struct {
			Error error
		}{Error: err}, "", "\t")
		w.Write(output)
		return nil, err
	}

	return users, nil
}

func getFilters(req *http.Request) map[string]string {
	result := make(map[string]string)
	req.ParseForm()
	for k, v := range req.Form {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}
