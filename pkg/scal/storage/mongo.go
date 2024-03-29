package storage

import (
	"errors"
	"io"

	"github.com/ilius/starcal-server/pkg/scal"

	mgo "github.com/ilius/mgo"
)

type hasHex interface {
	Hex() string
}

func Hex(id any) string {
	if id == nil {
		return ""
	}
	id2, ok := id.(hasHex)
	if !ok {
		log.Error("storage.Hex: can not convert to hex: ", id)
		return ""
	}
	return id2.Hex()
}

type MongoDatabase struct {
	mgo.Database
}

func (db *MongoDatabase) NewCondition(operator ConditionOperator) Condition {
	return &MongoCondition{
		op: operator,
	}
}

func (db *MongoDatabase) IsNotFound(err error) bool {
	return err == mgo.ErrNotFound
}

func (db *MongoDatabase) Insert(model hasCollection) error {
	return db.C(model.Collection()).Insert(model)
}

func (db *MongoDatabase) InsertMany(collection string, models []any) error {
	return db.C(collection).Insert(models...)
}

func (db *MongoDatabase) Update(model hasCollectionUniqueM) error {
	return db.C(model.Collection()).Update(
		model.UniqueM(),
		model,
	)
}

func (db *MongoDatabase) Upsert(model hasCollectionUniqueM) error {
	_, err := db.C(model.Collection()).Upsert(
		model.UniqueM(),
		model,
	)
	return err
}

func (db *MongoDatabase) Remove(model hasCollectionUniqueM) error {
	return db.C(model.Collection()).Remove(
		model.UniqueM(),
	)
}

func (db *MongoDatabase) RemoveAll(collection string, cond scal.M) (int, error) {
	info, err := db.C(collection).RemoveAll(
		cond,
	)
	count := 0
	if info != nil {
		count = info.Removed
	}
	return count, err
}

// func (db *MongoDatabase) Find(any)
func (db *MongoDatabase) Get(model hasCollectionUniqueM) error {
	return db.C(model.Collection()).Find(
		model.UniqueM(),
	).One(model)
}

func (db *MongoDatabase) First(
	cond scal.M,
	sortBy string,
	model hasCollection,
) error {
	return db.C(model.Collection()).Find(cond).Sort(sortBy).One(model)
}

func (db *MongoDatabase) FindCount(collection string, cond scal.M) (int, error) {
	return db.C(collection).Find(cond).Count()
}

func (db *MongoDatabase) FindAll(result any, in *FindInput) error {
	condition := in.Condition.Prepare()
	q := db.C(in.Collection).Find(condition)
	if in.SortBy != "" {
		sortBy := in.SortBy
		if in.ReverseOrder {
			sortBy = "-" + sortBy
		}
		q = q.Sort(sortBy)
	}
	if in.Limit > 0 {
		q = q.Limit(in.Limit)
	}
	if len(in.Fields) > 0 {
		selector := scal.M{}
		for _, field := range in.Fields {
			selector[field] = 1
		}
		q = q.Select(selector)
	}
	return q.All(result)
}

func (db *MongoDatabase) PipeIter(
	colName string,
	pipeline *[]scal.M,
) (func(any) error, func()) {
	iter := db.C(colName).Pipe(pipeline).Iter()
	return func(result any) error {
			if iter.Next(result) {
				return nil
			}
			if err := iter.Err(); err != nil {
				return err
			}
			if iter.Timeout() {
				return errors.New("timeout")
			}
			return io.EOF
		}, func() {
			err := iter.Close()
			if err != nil {
				log.Error("error in closing mongo iter: ", err)
			}
		}
}

// creates / checks index, panics on error
func (db *MongoDatabase) EnsureIndex(collection string, index mgo.Index) {
	err := db.Database.C(collection).EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

// creates / checks index with TTL, panics on error
func (db *MongoDatabase) EnsureIndexWithTTL(collection string, index mgo.Index) {
	err := EnsureIndexWithTTL(db.Database, collection, index)
	if err != nil {
		panic(err)
	}
}
