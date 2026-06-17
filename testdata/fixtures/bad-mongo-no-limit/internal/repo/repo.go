package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ListUsers(ctx context.Context, client *mongo.Client) error {
	coll := client.Database("app").Collection("users")
	_, err := coll.Find(ctx, bson.M{"active": true})
	return err
}
