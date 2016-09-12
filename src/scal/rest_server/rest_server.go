package rest_server

import (
    "fmt"
    "html"
    "log"
    "net/http"
    "encoding/json"

    //"gopkg.in/mgo.v2-unstable/bson"
    //"github.com/gorilla/mux"

)


func StartRestServer() {
    router := NewRouter()
    log.Fatal(http.ListenAndServe(":8080", router))
}

func Index(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
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
