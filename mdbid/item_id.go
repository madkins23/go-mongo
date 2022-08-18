package mdbid

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ObjectIDer provides an interface to items that use the primitive Mongo ObjectID.
type ObjectIDer interface {
	ID() primitive.ObjectID
	Filter() bson.D
}

// OIDmixin instantiates the ObjectIDer interface.
type OIDmixin struct {
	OID primitive.ObjectID `bson:"_id,omitempty"`
}

// ID returns the primitive Mongo ObjectID for an item.
func (idm *OIDmixin) ID() primitive.ObjectID {
	return idm.OID
}

// Filter returns a Mongo filter object for the item's ID.
func (idm *OIDmixin) Filter() bson.D {
	return bson.D{{"_id", idm.OID}}
}
