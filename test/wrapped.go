package test

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/madkins23/go-type/reg"
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
