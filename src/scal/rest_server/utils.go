package rest_server

import (
    "strings"
    "log"
    "net/url"
    "net/http"
    "encoding/json"
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


