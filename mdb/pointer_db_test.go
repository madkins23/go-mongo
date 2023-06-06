package mdb

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/madkins23/go-serial/pointer"

	"github.com/madkins23/go-mongo/mdbson"
)

type pointerTestSuite struct {
	AccessTestSuite
	showSerialized    bool
	targetCollection  *TypedCollection[TargetItem]
	pointerCollection *TypedCollection[PointerDemo]
}

func TestPointerSuite(t *testing.T) {
	suite.Run(t, new(pointerTestSuite))
}

func (suite *pointerTestSuite) SetupSuite() {
	if showSerialized, found := os.LookupEnv("GO-TYPE-SHOW-SERIALIZED"); found {
		var err error
		suite.showSerialized, err = strconv.ParseBool(showSerialized)
		suite.Require().NoError(err)
	}
	suite.AccessTestSuite.SetupSuite()
	suite.targetCollection = ConnectTypedCollectionHelper[TargetItem](
		&suite.AccessTestSuite, testCollectionValidation)
	suite.pointerCollection = ConnectTypedCollectionHelper[PointerDemo](
		&suite.AccessTestSuite, testCollection)
}

func (suite *pointerTestSuite) SetupTest() {
	pointer.ClearTargetCache()
	pointer.ClearFinderCache()
}

func (suite *pointerTestSuite) TearDownTest() {
	suite.NoError(suite.targetCollection.DeleteAll())
	suite.NoError(suite.pointerCollection.DeleteAll())
}

//////////////////////////////////////////////////////////////////////////

func (suite *pointerTestSuite) TestPointerFinder() {
	// This test will rely on automatic loading of the target cache
	// when the demo object is saved to the DB.
	// The Finder will not be used in this case.
	suite.testPointerFinder("finder", false)
}

func (suite *pointerTestSuite) TestPointerFinderWithDB() {
	// This test will wipe the target cache after saving the demo object.
	// The test case
	suite.testPointerFinder("database", true)
}

//------------------------------------------------------------------------

func (suite *pointerTestSuite) testPointerFinder(demoName string, forceFinder bool) {
	var targetItems = targetItems()

	// Load suite.targetCollection with records.
	for _, item := range targetItems {
		suite.Require().False(pointer.HasTarget(item.Group(), item.Key()))
		suite.Require().NoError(suite.targetCollection.Create(item))
	}

	// Add one record to the cache.
	suite.Require().NoError(pointer.SetTarget(targetItems[0], false))

	// Check to see if only the one item is in the target cache.
	for i, item := range targetItems {
		if i == 0 {
			suite.True(pointer.HasTarget(item.Group(), item.Key()))
		} else {
			suite.False(pointer.HasTarget(item.Group(), item.Key()))
		}
	}

	// Create PointerDemo object.
	demo := &PointerDemo{
		Name:   demoName,
		Single: mdbson.Point[*TargetItem](targetItems[0]),
		Array:  make([]*mdbson.Pointer[*TargetItem], 0),
		Map:    make(map[string]*mdbson.Pointer[*TargetItem]),
	}
	for _, item := range targetItems {
		demo.Array = append(demo.Array, mdbson.Point[*TargetItem](item))
		demo.Map[item.Charlie] = mdbson.Point[*TargetItem](item)
	}
	if suite.showSerialized {
		spew.Dump(demo)
	}

	// Put the demo object into suite.pointerCollection.
	// Note: This will fill the target cache as it goes along.
	// The items in the cache will be pre-store to the DB,
	// so they won't have Mongo ObjectIDs.
	suite.Require().NoError(suite.pointerCollection.Create(demo))

	finderCount := 0
	if forceFinder {
		// Clear the target cache to force the Finder to be called.
		pointer.ClearTargetCache()

		// Set a Finder in the target cache to pull targets from suite.targetCollection.
		suite.Require().NoError(pointer.SetFinder("simple", func(key string) (pointer.Target, error) {
			finderCount++
			item, err := suite.targetCollection.Find(bson.D{
				{"alpha", key},
			})
			if IsNotFound(err) {
				return nil, err
			} else if err != nil {
				return nil, fmt.Errorf("find record: %w", err)
			} else {
				return item, nil
			}
		}, false))
	}

	// Read the demo object back from suite.pointerCollection.
	readBack, err := suite.pointerCollection.Find(bson.D{{"name", demo.Name}})
	if suite.showSerialized {
		fmt.Println("-----------------------------")
		spew.Dump(readBack)
	}
	suite.Require().NoError(err)
	readBack.clearObjectIDs()
	suite.Equal(demo, readBack)

	if forceFinder {
		// Make sure Finder executed:
		suite.Len(targetItems, finderCount)

		// Note: In this mode the target cache has been rebuilt from the DB.
		// The TargetItem entries now have Mongo ObjectIDs
		// so they won't match the ones in the original demo object.
	} else {
		// Make sure the Finder did NOT execute:
		suite.Equal(0, finderCount)

		// The Pointer items should be singletons from the target cache.
		suite.True(demo.Single.Get() == readBack.Single.Get())
		for index, item := range demo.Array {
			suite.True(item.Get() == readBack.Array[index].Get())
		}
		for key, item := range demo.Map {
			suite.True(item.Get() == readBack.Map[key].Get())
		}
	}

	// Check for records in target cache, should all be present now.
	for _, item := range targetItems {
		suite.True(pointer.HasTarget(item.Group(), item.Key()))
	}
}

//========================================================================

type PointerDemo struct {
	Name   string
	Single *mdbson.Pointer[*TargetItem]
	Array  []*mdbson.Pointer[*TargetItem]
	Map    map[string]*mdbson.Pointer[*TargetItem]
}

func (pd *PointerDemo) clearObjectIDs() {
	pd.Single.Get().clearID()
	for _, ptr := range pd.Array {
		ptr.Get().clearID()
	}
	for _, ptr := range pd.Map {
		ptr.Get().clearID()
	}
}

var _ pointer.Target = &TargetItem{}

type TargetItem struct {
	SimpleItem `bson:"inline"`
}

func (ti *TargetItem) clearID() {
	ti.ObjectID = [12]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
}

func (ti *TargetItem) Group() string {
	return "simple"
}

func (ti *TargetItem) Key() string {
	return ti.Alpha
}

var tgtItems []*TargetItem

func targetItems() []*TargetItem {
	if tgtItems == nil {
		for _, simpleItem := range simpleItems {
			tgtItems = append(tgtItems, &TargetItem{SimpleItem: *simpleItem})
		}
	}
	return tgtItems
}

var simpleItems = []*SimpleItem{
	SimpleItem1, SimpleItem1x, SimpleItem2, SimpleItem3, UnfilteredItem,
}
