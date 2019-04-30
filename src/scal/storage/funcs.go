package storage

import (
	"reflect"
	"scal"
)

func PipeAll(db Database, collection string, pipeline *[]scal.M, result interface{}) error {
	next, close := db.PipeIter(collection, pipeline)
	defer close()

	resultv := reflect.ValueOf(result)
	if resultv.Kind() != reflect.Ptr {
		panic("result argument must be a slice address")
	}

	slicev := resultv.Elem()

	if slicev.Kind() == reflect.Interface {
		slicev = slicev.Elem()
	}
	if slicev.Kind() != reflect.Slice {
		panic("result argument must be a slice address")
	}

	slicev = slicev.Slice(0, slicev.Cap())
	elemt := slicev.Type().Elem()
	i := 0
	for {
		if slicev.Len() == i {
			elemp := reflect.New(elemt)
			err := next(elemp.Interface())
			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				return err
			}
			slicev = reflect.Append(slicev, elemp.Elem())
			slicev = slicev.Slice(0, slicev.Cap())
		} else {
			err := next(slicev.Index(i).Addr().Interface())
			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				return err
			}
		}
		i++
	}
	resultv.Elem().Set(slicev.Slice(0, i))
	return nil
}
