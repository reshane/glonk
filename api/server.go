package api

import (
    "net/http"
    "encoding/json"
    "strconv"
    "log"

    "github.com/gorilla/mux"

    "github.com/reshane/glonk/store"
    "github.com/reshane/glonk/types"
)

type Server struct {
    listenAddr string
    db store.Store
}

func NewServer(listenAddr string, db store.Store) *Server {
    return &Server {
        listenAddr: listenAddr,
        db: db,
    }
}

func (s *Server) Start() error {
    r := mux.NewRouter()
    // data endpoints
    r.Handle("/data/{dataType}/{id}", isAuthorized(s.handleGetByID)).
        Methods("GET")
    r.Handle("/data/{dataType}", isAuthorized(s.handleCreate)).
        Methods("POST")
    r.Handle("/data/{dataType}", isAuthorized(s.handleUpdate)).
        Methods("PUT")
    r.Handle("/data/{dataType}/{id}", isAuthorized(s.handleDeleteByID)).
        Methods("DELETE")

    // auth
    r.HandleFunc("/auth/google/login", s.googleLogin)
    r.HandleFunc("/auth/google/callback", s.googleCallback)
    r.Handle("/", http.FileServer(http.Dir("./templates")))

    http.Handle("/", r)
    return http.ListenAndServe(s.listenAddr, nil)
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    dataType := vars["dataType"]
    if _, exists := types.TypeStrings[dataType]; !exists {
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }

    data, err := types.Decoders[dataType](r)
    if err != nil {
        log.Println(err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    if !data.Validate() {
        log.Println("Invalid data in update request:", data)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    updated, err := s.db.Update(data)
    if err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    json.NewEncoder(w).Encode(updated)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    dataType := vars["dataType"]
    if _, exists := types.TypeStrings[dataType]; !exists {
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }

    data, err := types.Decoders[dataType](r)
    if err != nil {
        log.Println(err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    if !data.Validate() {
        log.Println("Invalid data in request:", data)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    created, err := s.db.Create(data)
    if err != nil {
        log.Println("Could not create object:", err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    json.NewEncoder(w).Encode(created)
}

func (s *Server) handleDeleteByID(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    dataType := vars["dataType"]
    if _, exists := types.TypeStrings[dataType]; !exists {
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }

    id, err := strconv.ParseInt(vars["id"], 10, 64)
    if err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }


    err = s.db.Delete(dataType, id)
    if err != nil {
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }
    w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGetByID(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    dataType := vars["dataType"]
    if _, exists := types.TypeStrings[dataType]; !exists {
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }

    id, err := strconv.ParseInt(vars["id"], 10, 64)
    if err != nil {
        log.Println("Could not parse int from id: ", vars["id"], err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    data, err := s.db.Get(dataType, id)
    if err != nil {
        log.Println("Could not find data:", err)
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(data)
}
