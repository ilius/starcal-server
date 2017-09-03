package api_v1

import (
	"fmt"
	"net/url"
	"strings"

	. "github.com/ilius/restpc"
	"gopkg.in/mgo.v2/bson"
)

func SplitURL(u *url.URL) []string {
	return strings.Split(strings.Trim(u.Path, "/"), "/")
}

func ObjectIdFromURL(req Request, name string, indexFromEnd int) (*bson.ObjectId, error) {
	u := req.URL()
	parts := SplitURL(u)
	if len(parts) < 2 {
		return nil, NewError(Internal, "", fmt.Errorf("Unexpected URL: %s", u))
	}
	objIdHex := parts[len(parts)-1-indexFromEnd]
	if !bson.IsObjectIdHex(objIdHex) { // to avoid panic!
		return nil, NewError(InvalidArgument, fmt.Sprintf("invalid '%s'", name), nil)
	}
	objId := bson.ObjectIdHex(objIdHex)
	return &objId, nil
}
