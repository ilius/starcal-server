package rest_server

import "log"
import "net/http"
import _ "scal/init/cal_types_init"

func StartEventRestServer() {
    router := GetEventRouter()
    log.Fatal(http.ListenAndServe(":8080", router))
}

