package mdb

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IndexDescription struct {
	unique bool
	keys   []string
}

// NewIndexDescription creates a new index description.
func NewIndexDescription(unique bool, keys ...string) *IndexDescription {
	return &IndexDescription{
		unique: unique,
		keys:   keys,
	}
}

func (id *IndexDescription) AsBSON() bson.D {
	asBSON := bson.D{}
	for _, key := range id.keys {
		asBSON = append(asBSON, bson.E{Key: key, Value: 1})
	}
	return asBSON
}

// Finisher returns a function that can be used as a CollectionFinisher for creating this index.
func (id *IndexDescription) Finisher() CollectionFinisher {
	return func(access *Access, collection *Collection) error {
		return access.Index(collection, id)
	}
}

func (a *Access) Index(collection *Collection, description *IndexDescription) error {
	ctx, cancel := a.ContextWithTimeout(a.config.Timeout.Index)
	defer cancel()
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: description.AsBSON(),
		Options: &options.IndexOptions{
			Unique: &description.unique,
		},
	})
	if err != nil {
		// TODO(mAdkins): at this point should the index be removed?
		//  Experimentation suggests that double creation of the index is OK.
		return fmt.Errorf("create index on name: %w", err)
	}

	a.Info("Created index on collection " + collection.Name())

	return nil
}
