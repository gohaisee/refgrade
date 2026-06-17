package repo

import "go.mongodb.org/mongo-driver/bson"

func Filter(input string) bson.M {
	return bson.M{"$where": input}
}
