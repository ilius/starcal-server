package api_v1

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/mgo.v2/bson"
)

func SplitURL(u *url.URL) []string {
	return strings.Split(strings.Trim(u.Path, "/"), "/")
}

func SetHttpError(w http.ResponseWriter, code int, msg string) {
	if code == http.StatusInternalServerError {
		log.Print("Internal Server Error: ", msg)
		// don't expose internal error messages to outsiders
		msg = "Internal Server Error"
	}
	jsonByte, _ := json.Marshal(map[string]string{
		"error": msg,
	})
	http.Error(
		w,
		string(jsonByte),
		code,
	)
}

func SetHttpErrorInternalMsg(w http.ResponseWriter, errMsg string) {
	SetHttpError(w, http.StatusInternalServerError, errMsg)
}

func SetHttpErrorInternal(w http.ResponseWriter, err error) {
	SetHttpError(w, http.StatusInternalServerError, err.Error())
}

func SetHttpErrorUserNotFound(w http.ResponseWriter, email string) {
	SetHttpError(
		w,
		http.StatusInternalServerError,
		fmt.Sprintf(
			"user with email '%s' not found",
			email,
		),
	)
}

func ObjectIdFromURL(
	w http.ResponseWriter,
	r *http.Request,
	name string,
	indexFromEnd int,
) *bson.ObjectId {
	parts := SplitURL(r.URL)
	if len(parts) < 2 {
		SetHttpErrorInternalMsg(
			w,
			fmt.Sprintf(
				"Unexpected URL: %s",
				r.URL,
			),
		)
		return nil
	}
	objIdHex := parts[len(parts)-1-indexFromEnd]
	if !bson.IsObjectIdHex(objIdHex) { // to avoid panic!
		SetHttpError(
			w,
			http.StatusBadRequest,
			fmt.Sprintf(
				"invalid '%s'",
				name,
			),
		)
		return nil
	}
	objId := bson.ObjectIdHex(objIdHex)
	return &objId
}
