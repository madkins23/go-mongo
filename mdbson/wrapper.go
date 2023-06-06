package mdbson

import (
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"

	"github.com/madkins23/go-type/reg"
)

// Wrap an item in a BSON wrapper that can handle serialization.
func Wrap[W any](item W) *Wrapper[W] {
	w := new(Wrapper[W])
	w.Set(item)
	return w
}

// Wrapper is used to attach a type name to an item to be serialized.
// This supports re-creating the correct type for filling an interface field.
type Wrapper[T any] struct {
	item T
}

// Get the wrapped item.
func (w *Wrapper[T]) Get() T {
	return w.item
}

// Set the wrapped item.
func (w *Wrapper[T]) Set(t T) {
	w.item = t
}

type packed struct {
	TypeName string   `bson:"type"`
	RawForm  bson.Raw `bson:"data"`
}

func (w *Wrapper[T]) MarshalBSON() ([]byte, error) {
	var err error
	var pack packed
	if pack.TypeName, err = reg.NameFor(w.item); err != nil {
		return nil, fmt.Errorf("get type name for %#v: %w", w.item, err)
	}

	build := &strings.Builder{}
	var writer bsonrw.ValueWriter
	var encoder *bson.Encoder
	if writer, err = bsonrw.NewBSONValueWriter(build); err != nil {
		return nil, fmt.Errorf("create BSON value writer: %w", err)
	} else if encoder, err = bson.NewEncoder(writer); err != nil {
		return nil, fmt.Errorf("create BSON encoder: %w", err)
	} else if err = encoder.Encode(w.item); err != nil {
		return nil, fmt.Errorf("marshal packed area: %w", err)
	}
	// Must get rid of extraneous ending newline that is not unmarshaled.
	pack.RawForm = []byte(strings.TrimSuffix(build.String(), "\n"))

	var marshaled []byte
	marshaled, err = bson.Marshal(pack)
	if err != nil {
		return []byte(""), fmt.Errorf("marshal packed form: %w", err)
	}
	return marshaled, nil
}

func (w *Wrapper[T]) UnmarshalBSON(marshaled []byte) error {
	var pack packed
	if err := bson.Unmarshal(marshaled, &pack); err != nil {
		return fmt.Errorf("unmarshal packed area: %w", err)
	}

	var ok bool
	var decoder *bson.Decoder
	if pack.TypeName == "" {
		return fmt.Errorf("empty type field")
	} else if temp, err := reg.Make(pack.TypeName); err != nil {
		return fmt.Errorf("make instance of type %s: %w", pack.TypeName, err)
	} else if decoder, err = bson.NewDecoder(bsonrw.NewBSONDocumentReader(pack.RawForm)); err != nil {
		return fmt.Errorf("create BSON decoder: %w", err)
	} else if err = decoder.Decode(temp); err != nil {
		return fmt.Errorf("decode wrapper contents: %w", err)
	} else if w.item, ok = temp.(T); !ok {
		// TODO(mAdkins): How to get name of T?
		return fmt.Errorf("type %s not generic type", pack.TypeName)
	} else {
		return nil
	}
}
