package storage

import (
	"scal"
)

type hasCollection interface {
	Collection() string
}

type hasCollectionUniqueM interface {
	Collection() string
	UniqueM() scal.M
}

type FindInput struct {
	Collection string
	Conditions scal.M
	SortBy     string
	Limit      int
	Fields     []string
}

type Database interface {
	IsNotFound(error) bool
	Insert(model hasCollection) error
	InsertMany(collection string, models []interface{}) error
	Update(model hasCollectionUniqueM) error
	Upsert(model hasCollectionUniqueM) error
	Remove(model hasCollectionUniqueM) error
	Get(model hasCollectionUniqueM) error
	First(scal.M, string, hasCollection) error
	FindCount(string, scal.M) (int, error)
	FindAll(result interface{}, input *FindInput) error
	PipeAll(collection string, pipeline *[]scal.M, result interface{}) error
	PipeIter(collection string, pipeline *[]scal.M) (next func() (scal.M, error), close func())
}
