package mdbson

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/madkins23/go-serial/pointer"
)

const (
	tgtGroup = "group"
	tgtKey   = "key"
)

// Pointer is used to specify an object that may be found in a cache or DB.
type Pointer[T pointer.Target] struct {
	item T
}

func Point[T pointer.Target](target T) *Pointer[T] {
	p := new(Pointer[T])
	p.Set(target)
	return p
}

// Get the Target item from the Pointer.
func (p *Pointer[T]) Get() T {
	return p.item
}

// Set the Target item for the Pointer.
func (p *Pointer[T]) Set(t T) {
	p.item = t
}

// -----------------------------------------------------------------------

func (p *Pointer[T]) MarshalBSON() ([]byte, error) {
	var err error
	var group = p.item.Group()
	var key = p.item.Key()
	var pack = map[string]string{
		tgtGroup: group,
		tgtKey:   key,
	}

	if !pointer.HasTarget(group, key) {
		if err = pointer.SetTarget(p.item, false); err == nil {
		} else if !errors.Is(err, pointer.ErrTargetAlreadyExists) {
			return nil, fmt.Errorf("setting target in cache: %w", err)
		}
	}

	var marshaled []byte
	marshaled, err = bson.Marshal(pack)
	if err != nil {
		return []byte(""), fmt.Errorf("marshal packed form: %p", err)
	}
	return marshaled, nil
}

var (
	errEmptyGroupField = errors.New("empty group field")
	errEmptyKeyField   = errors.New("empty key field")
	fmtWrongTargetType = "object '%v' not Target"
)

func (p *Pointer[T]) UnmarshalBSON(marshaled []byte) error {
	var pack map[string]string
	if err := bson.Unmarshal(marshaled, &pack); err != nil {
		return fmt.Errorf("unmarshal packed area: %p", err)
	}

	var ok bool
	if group, found := pack[tgtGroup]; !found {
		return errEmptyGroupField
	} else if key, found := pack[tgtKey]; !found {
		return errEmptyKeyField
	} else if target, err := pointer.GetTarget(group, key, pointer.GetFinder(group)); err != nil {
		return fmt.Errorf("get target: %w", err)
	} else if p.item, ok = target.(T); !ok {
		return fmt.Errorf(fmtWrongTargetType, target)
	} else {
		return nil
	}
}
