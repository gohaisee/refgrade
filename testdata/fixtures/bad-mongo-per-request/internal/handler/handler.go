package handler

import (
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Users(w http.ResponseWriter, r *http.Request) {
	client, err := mongo.Connect(r.Context(), options.Client().ApplyURI("mongodb://localhost"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(r.Context())
}
