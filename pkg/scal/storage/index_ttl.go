package storage

import (
	"fmt"
	"strings"
	"time"

	mgo "github.com/ilius/mgo"
	"github.com/ilius/mgo/bson"
)

func ModifyIndexTTL(db mgo.Database, collection string, index mgo.Index) error {
	keyInfo, err := mgo.ParseIndexKey(index.Key)
	if err != nil {
		return err
	}
	expireAfterSeconds := int(index.ExpireAfter / time.Second)
	fmt.Printf(
		"Updating TTL on collection %s to expireAfterSeconds=%d\n",
		collection,
		expireAfterSeconds,
	)
	err = db.Run(bson.D{
		{"collMod", collection},
		{"index", bson.M{
			"keyPattern":         keyInfo.Key,
			"expireAfterSeconds": expireAfterSeconds,
		}},
	}, nil)
	if err != nil {
		return err
	}
	return nil
}

func EnsureIndexWithTTL(db mgo.Database, collection string, index mgo.Index) error {
	err := db.C(collection).EnsureIndex(index)
	if err != nil {
		// if expireAfterSeconds is changed, we need to drop and re-create the index
		// unless we use `collMod` added in 2.3.2
		// https://jira.mongodb.org/browse/SERVER-6700
		if strings.Contains(err.Error(), "already exists with different options") {
			return ModifyIndexTTL(db, collection, index)
		}
		return err
	}
	return nil
}
