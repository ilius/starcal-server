package rest_server

import (
    "fmt"
    "strings"
    "log"
    "net/url"
    "net/http"
    "encoding/json"

    "gopkg.in/mgo.v2-unstable/bson"
    "scal-lib/go-http-auth"
)

func SplitURL(u *url.URL) []string {
    return strings.Split(strings.Trim(u.Path, "/"), "/")
}

func SetHttpError(w http.ResponseWriter, code int, msg string){
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

func ObjectIdFromURL(
    w http.ResponseWriter,
    r *auth.AuthenticatedRequest,
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
    if !bson.IsObjectIdHex(objIdHex) {// to avoid panic!
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
