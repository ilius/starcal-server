package storage

import (
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

type hasCollection interface {
    Collection() string
}

type hasCollectionUniqueM interface {
    Collection() string
    UniqueM() bson.M
}

func Insert(db *mgo.Database, model hasCollection) error {
    return db.C(model.Collection()).Insert(model)
}

func Update(db *mgo.Database, model hasCollectionUniqueM) error {
    return db.C(model.Collection()).Update(
        model.UniqueM(),
        model,
    )
}

func Upsert(db *mgo.Database, model hasCollectionUniqueM) error {
    _, err := db.C(model.Collection()).Upsert(
        model.UniqueM(),
        model,
    )
    return err
}

func Remove(db *mgo.Database, model hasCollectionUniqueM) error {
    return db.C(model.Collection()).Remove(
        model.UniqueM(),
    )
}



