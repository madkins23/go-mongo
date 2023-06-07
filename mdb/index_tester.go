package mdb

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IndexTester provides a utility for verifying index creation.
type IndexTester []indexDatum

type indexDatum struct {
	Name   string
	Key    map[string]int32
	Unique bool
}

func NewIndexTester() IndexTester {
	return make(IndexTester, 0, 2)
}

func (it IndexTester) TestIndexes(t *testing.T, collection *Collection, descriptions ...*IndexDescription) {
	ctx := context.Background()
	cursor, err := collection.Indexes().List(ctx)
	require.NoError(t, err)
	err = cursor.All(ctx, &it)
	require.NoError(t, err)
	assert.Len(t, it, len(descriptions)+1)
	it.hasIndexNamed(t, "_id_", NewIndexDescription(false, "_id"))
	for _, description := range descriptions {
		nameMap := make([]string, 0, len(description.keys))
		for _, key := range description.keys {
			nameMap = append(nameMap, key+"_1")
		}
		it.hasIndexNamed(t, strings.Join(nameMap, "_"), description)
	}
}

func (it IndexTester) hasIndexNamed(t *testing.T, name string, description *IndexDescription) {
	for _, data := range it {
		if data.Name == name {
			assert.Equal(t, description.unique, data.Unique, "check unique for index %s", name)
			keyMap := make(map[string]int32, len(description.keys))
			for _, key := range description.keys {
				keyMap[key] = 1
			}
			assert.Equal(t, keyMap, data.Key, "check keys for index %s", name)
			return
		}
	}

	names := make([]string, 0, len(it))
	for _, data := range it {
		names = append(names, data.Name)
	}
	assert.Fail(t, "missing index", "no index %s (%s)", name, strings.Join(names, ", "))
}
