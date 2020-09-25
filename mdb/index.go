package mdb

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (a *Access) Index(collection *mongo.Collection, fields bson.D, unique bool) error {
	ctx, cancel := a.ContextWithTimeout(a.config.Timeout.Index)
	defer cancel()
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{"name", 1},
		},
		Options: &options.IndexOptions{
			Unique: &unique, // magic cookie: must be address of boolean
		},
	})
	if err != nil {
		// TODO: at this point should the collection be removed?
		// Experimentation suggests that double creation of the index is OK.
		return fmt.Errorf("create index on name: %w", err)
	}

	a.Info("Created index on collection " + collection.Name())

	return nil
}
