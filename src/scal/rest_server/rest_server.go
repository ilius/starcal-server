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
    jsonByte, _ := json.Marshal(map[string]string{
        "error": msg,
    })
    http.Error(
        w,
        string(jsonByte),
        code,
    )
}



/*
func EventsList(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=UTF-8")
    db, err_db := storage.GetDB()
    if err_db != nil {
        http.Error(w, err_db.Error(), http.StatusInternalServerError)
        return
    }
    var results []event_lib.BaseEventModel
    err_results := db.C("events").Find(bson.M{"ownerId": 0}).Sort("_id").All(&results)
    if err_results != nil {
        http.Error(w, err_results.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(results)
}*/





