package mdb

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/madkins23/go-type/reg"

	"github.com/madkins23/go-mongo/mdbson"
)

// RegisterWrapped registers structs that will be wrapped during testing.
// Uses the github.com/madkins23/go-type library to register structs by name.
func RegisterWrapped() error {
	if err := reg.Register(&TextValue{}); err != nil {
		return fmt.Errorf("registering TextValue struct: %w", err)
	}
	if err := reg.Register(&NumericValue{}); err != nil {
		return fmt.Errorf("registering NumericValue struct: %w", err)
	}
	if err := reg.Register(&RandomValue{}); err != nil {
		return fmt.Errorf("registering RandomValue struct: %w", err)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type Wrappable interface {
	Key() string
	fmt.Stringer
}

const (
	ValueText     = "lorem ipsum"
	ValueNumber   = 123
	RandomMinimum = 8
	RandomRandom  = 8
	RandomMaximum = RandomMinimum + RandomRandom
)

var _ Wrappable = &TextValue{}

type TextValue struct {
	Text string
}

func (sv *TextValue) Key() string {
	return "text"
}

func (sv *TextValue) String() string {
	return sv.Text
}

var _ Wrappable = &NumericValue{}

type NumericValue struct {
	Number int
}

func (nv *NumericValue) Key() string {
	return "numeric"
}

func (nv *NumericValue) String() string {
	return strconv.Itoa(nv.Number)
}

var _ Wrappable = &RandomValue{}

type RandomValue struct {
}

func (rv *RandomValue) Key() string {
	return "random"
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (rv *RandomValue) String() string {
	b := make([]rune, RandomMinimum+rand.Intn(RandomRandom))
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

////////////////////////////////////////////////////////////////////////////////

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
