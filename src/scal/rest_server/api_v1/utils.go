package api_v1

import (
	"fmt"

	. "github.com/ilius/restpc"
	"gopkg.in/mgo.v2/bson"
)

func ObjectIdFromURL(req Request, name string, indexFromEnd int) (*bson.ObjectId, error) {
	objIdHex, err := req.GetString(name)
	if err != nil {
		return nil, err
	}
	if !bson.IsObjectIdHex(*objIdHex) { // to avoid panic!
		return nil, NewError(InvalidArgument, fmt.Sprintf("invalid '%s'", name), nil)
	}
	objId := bson.ObjectIdHex(*objIdHex)
	return &objId, nil
}
