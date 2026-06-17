package handler

import "go.mongodb.org/mongo-driver/bson"

type userDoc struct {
	ID   string `bson:"_id"`
	Name string `bson:"name"`
}

func Decode(raw []byte) (userDoc, error) {
	var doc userDoc
	err := bson.Unmarshal(raw, &doc)
	return doc, err
}
