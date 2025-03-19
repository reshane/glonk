package api

import (
    "net/http"
    "encoding/json"
    "strconv"

    "github.com/gorilla/mux"

    "github.com/reshane/gorest/store"
    "github.com/reshane/gorest/types"
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
    r.Handle("/{dataType}/{id}", isAuthorized(s.handleGetByID)).
        Methods("GET")
    r.Handle("/{dataType}", isAuthorized(s.handleCreate)).
        Methods("POST")
    r.Handle("/{dataType}", isAuthorized(s.handleUpdate)).
        Methods("PUT")
    r.Handle("/{dataType}/{id}", isAuthorized(s.handleDeleteByID)).
        Methods("DELETE")
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
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    if !data.Validate() {
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
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    if !data.Validate() {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    created, err := s.db.Create(data)
    if err != nil {
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

    id, err := strconv.Atoi(vars["id"])
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

    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    data, err := s.db.Get(dataType, id)
    if err != nil {
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }
    json.NewEncoder(w).Encode(data)
}
