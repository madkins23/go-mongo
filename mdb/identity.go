package mdb

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Identifier provides an interface to items that use the primitive Mongo ObjectID.
type Identifier interface {
	ID() primitive.ObjectID
	IDfilter() bson.D
}

// Identity instantiates the Identifier interface.
type Identity struct {
	ObjectID primitive.ObjectID `bson:"_id,omitempty"`
}

// ID returns the primitive Mongo ObjectID for an item.
func (idm *Identity) ID() primitive.ObjectID {
	return idm.ObjectID
}

// Filter returns a Mongo filter object for the item's ID.
func (idm *Identity) IDfilter() bson.D {
	return bson.D{{"_id", idm.ObjectID}}
}
