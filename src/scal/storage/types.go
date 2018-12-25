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
	Cond       scal.M
	SortBy     string
}

type Database interface {
	IsNotFound(error) bool
	Insert(model hasCollection) error
	Update(model hasCollectionUniqueM) error
	Upsert(model hasCollectionUniqueM) error
	Remove(model hasCollectionUniqueM) error
	Get(model hasCollectionUniqueM) error
	First(scal.M, string, hasCollection) error
	FindCount(string, scal.M) (int, error)
	FindAll(result interface{}, input *FindInput) error
	PipeAll(string, *[]scal.M, interface{}) error
	PipeIter(string, *[]scal.M) <-chan scal.MErr
}
