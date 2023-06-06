//go:build database

package mdb

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type indexTestSuite struct {
	AccessTestSuite
	collection *Collection
}

func TestIndexSuite(t *testing.T) {
	suite.Run(t, new(indexTestSuite))
}

func (suite *indexTestSuite) SetupTest() {
	suite.collection = suite.ConnectCollection(testCollectionValidation)
}

func (suite *indexTestSuite) TearDownTest() {
	_ = suite.collection.Drop()
}

func (suite *indexTestSuite) TestIndexNone() {
	NewIndexTester().TestIndexes(suite.T(), suite.collection)
}

func (suite *indexTestSuite) TestIndexOne() {
	index1 := NewIndexDescription(true, "alpha")
	suite.Require().NoError(suite.access.Index(suite.collection, index1))
	NewIndexTester().TestIndexes(suite.T(), suite.collection, index1)
}

func (suite *indexTestSuite) TestIndexTwo() {
	index1 := NewIndexDescription(true, "alpha")
	index2 := NewIndexDescription(true, "bravo")
	suite.Require().NoError(suite.access.Index(suite.collection, index1))
	suite.Require().NoError(suite.access.Index(suite.collection, index2))
	NewIndexTester().TestIndexes(suite.T(), suite.collection, index1, index2)
}

func (suite *indexTestSuite) TestIndexThree() {
	index1 := NewIndexDescription(true, "alpha")
	index2 := NewIndexDescription(true, "bravo")
	index3 := NewIndexDescription(true, "alpha", "bravo")
	suite.Require().NoError(suite.access.Index(suite.collection, index1))
	suite.Require().NoError(suite.access.Index(suite.collection, index2))
	suite.Require().NoError(suite.access.Index(suite.collection, index3))
	NewIndexTester().TestIndexes(suite.T(), suite.collection, index1, index2, index3)
}

func (suite *indexTestSuite) TestIndexFinisher() {
	index := NewIndexDescription(true, "alpha", "bravo")
	collection, err := ConnectCollection(suite.access,
		&CollectionDefinition{
			Name:           "test-collection-index-finisher",
			ValidationJSON: SimpleValidatorJSON,
			Finishers: []CollectionFinisher{
				index.Finisher(),
			},
		})
	suite.Require().NoError(err)
	suite.NotNil(collection)
	NewIndexTester().TestIndexes(suite.T(), collection, index)
}

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

// =============================================================================

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
