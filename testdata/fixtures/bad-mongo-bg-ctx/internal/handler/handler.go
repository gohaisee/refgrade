package handler

import (
	"context"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func Users(w http.ResponseWriter, r *http.Request, client *mongo.Client) {
	coll := client.Database("app").Collection("users")
	_, _ = coll.Find(context.Background(), bson.M{})
}
