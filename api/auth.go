package api

import (
    // "log"
    "net/http"
    // "encoding/json"
)

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        endpoint(w, r)
        /*
        if r.Header["Authorization"] != nil {
            endpoint(w, r)
        } else {
            json.NewEncoder(w).Encode(map[string]string{"error": "Not Authorized"})
        }
        */
    })
}
