package storage

import (
	"gopkg.in/mgo.v2/bson"
)

type hasCollection interface {
	Collection() string
}

type hasCollectionUniqueM interface {
	Collection() string
	UniqueM() bson.M
}

