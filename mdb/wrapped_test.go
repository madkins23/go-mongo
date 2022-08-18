package mdb

import (
	"github.com/madkins23/go-mongo/mdbson"
)

// These items can't be defined in the test package due to an import cycle.
// They require the mdbson package which has tests which in turn require the test package.
// In addition, they are used by both typed_collection_db_test and cached_collection_db_

type WrappedItems struct {
	SimpleItem `bson:"inline"`
	Single     *mdbson.Wrapper[Wrappable]
	Array      []*mdbson.Wrapper[Wrappable]
	Map        map[string]*mdbson.Wrapper[Wrappable]
}

////////////////////////////////////////////////////////////////////////////////

func MakeWrappedItems() *WrappedItems {
	items := []*mdbson.Wrapper[Wrappable]{
		mdbson.Wrap[Wrappable](&TextValue{Text: ValueText}),
		mdbson.Wrap[Wrappable](&NumericValue{Number: ValueNumber}),
		mdbson.Wrap[Wrappable](&RandomValue{}),
	}
	wrapped := &WrappedItems{
		SimpleItem: SimpleItem{
			Alpha:   "Wrapped",
			Bravo:   23,
			Charlie: "Need this to pass validation",
		},
		Single: items[0],
		Array:  items,
		Map:    make(map[string]*mdbson.Wrapper[Wrappable], len(items)),
	}
	for _, item := range items {
		wrapped.Map[item.Get().Key()] = item
	}
	return wrapped
}
