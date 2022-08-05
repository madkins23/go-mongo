package mdb

import (
	"github.com/madkins23/go-mongo/mdbson"
	"github.com/madkins23/go-mongo/test"
)

// These items can't be defined in the test package due to an import cycle.
// They require the mdbson package which has tests which in turn require the test package.
// In addition, they are used by both typed_collection_db_test and cached_collection_db_test.

type WrappedItems struct {
	test.SimpleItem `bson:"inline"`
	Single          *mdbson.Wrapper[test.Wrappable]
	Array           []*mdbson.Wrapper[test.Wrappable]
	Map             map[string]*mdbson.Wrapper[test.Wrappable]
}

////////////////////////////////////////////////////////////////////////////////

func MakeWrappedItems() *WrappedItems {
	items := []*mdbson.Wrapper[test.Wrappable]{
		mdbson.Wrap[test.Wrappable](&test.TextValue{Text: test.ValueText}),
		mdbson.Wrap[test.Wrappable](&test.NumericValue{Number: test.ValueNumber}),
		mdbson.Wrap[test.Wrappable](&test.RandomValue{}),
	}
	wrapped := &WrappedItems{
		SimpleItem: test.SimpleItem{
			SimpleKey: test.SimpleKey{
				Alpha: "Wrapped",
				Bravo: 23,
			},
			Charlie: "Need this to pass validation",
		},
		Single: items[0],
		Array:  items,
		Map:    make(map[string]*mdbson.Wrapper[test.Wrappable], len(items)),
	}
	for _, item := range items {
		wrapped.Map[item.Get().Key()] = item
	}
	return wrapped
}
