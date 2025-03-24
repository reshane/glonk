package api

import (
    "net/http"
    "encoding/json"
    "strconv"
    "log"
    "fmt"

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
    r.Handle("/data/{dataType}", isAuthorized(s.handleGetByQueries)).
        Methods("GET")
    r.Handle("/data/{dataType}", isAuthorized(s.handleCreate)).
        Methods("POST")
    r.Handle("/data/{dataType}", isAuthorized(s.handleUpdate)).
        Methods("PUT")
    r.Handle("/data/{dataType}/{id}", isAuthorized(s.handleDeleteByID)).
        Methods("DELETE")

    // schema
    r.Handle("/schema", isAuthorized(s.schema)).
        Methods("GET")

    // auth
    r.HandleFunc("/auth/google/login", s.googleLogin)
    r.HandleFunc("/auth/google/callback", s.googleCallback)
    r.Handle("/", http.FileServer(http.Dir("./templates")))

    http.Handle("/", r)
    return http.ListenAndServe(s.listenAddr, nil)
}

func getOwnerIdFromRequestHeaders(r *http.Request) (int64, error) {
    ownerIdHeaderVal := r.Header.Get("OwnerID")
    if len(ownerIdHeaderVal) < 1 {
        return -1, fmt.Errorf("Invalid OwnerID Header value %s", ownerIdHeaderVal)
    }

    ownerId, err := strconv.ParseInt(ownerIdHeaderVal, 10, 64)
    if err != nil {
        return -1, err
    }
    return ownerId, nil
}

func (s *Server) schema(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(types.MetaDataMap)
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
    ownerId, err := getOwnerIdFromRequestHeaders(r)
    if err != nil {
        log.Println("Could not get ownerId from headers", err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    vars := mux.Vars(r)
    dataType := vars["dataType"]
    metaData, exists := types.MetaDataMap[dataType]
    if !exists {
        log.Println("No metaData for specified data type:", dataType)
        http.Error(w, "Not Found", http.StatusBadRequest)
        return
    }

    data, err := metaData.GetDecoder()(r)
    if err != nil {
        log.Println(err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    if data.GetOwnerId() != ownerId {
        log.Println("Session ownerId does not match data ownerId")
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
        log.Println(err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(updated)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
    ownerId, err := getOwnerIdFromRequestHeaders(r)
    if err != nil {
        log.Println("Could not get ownerId from headers", err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    vars := mux.Vars(r)
    dataType := vars["dataType"]
    metaData, exists := types.MetaDataMap[dataType]
    if !exists {
        log.Println("No metaData for specified data type:", dataType)
        http.Error(w, "Not Found", http.StatusBadRequest)
        return
    }

    data, err := metaData.GetDecoder()(r)
    if err != nil {
        log.Println(err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    if data.GetOwnerId() != ownerId {
        log.Println("Session ownerId does not match data ownerId")
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

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(created)
}

func (s *Server) handleDeleteByID(w http.ResponseWriter, r *http.Request) {
    ownerId, err := getOwnerIdFromRequestHeaders(r)
    if err != nil {
        log.Println("Could not get ownerId from headers", err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    vars := mux.Vars(r)
    dataType := vars["dataType"]
    metaData, exists := types.MetaDataMap[dataType]
    if !exists {
        log.Println("No metaData for specified data type:", dataType)
        http.Error(w, "Not Found", http.StatusBadRequest)
        return
    }

    id, err := strconv.ParseInt(vars["id"], 10, 64)
    if err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }


    data, err := s.db.Delete(metaData, id, ownerId)
    if err != nil {
        log.Println("Could not find data:", err)
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}

func (s *Server) handleGetByQueries(w http.ResponseWriter, r *http.Request) {
    ownerId, err := getOwnerIdFromRequestHeaders(r)
    if err != nil {
        log.Println("Could not get ownerId from headers", err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    vars := mux.Vars(r)
    dataType := vars["dataType"]
    metaData, exists := types.MetaDataMap[dataType]
    if !exists {
        log.Println("No metaData for specified data type:", dataType)
        http.Error(w, "Not Found", http.StatusBadRequest)
        return
    }

    parsers := metaData.GetQueries()
    queries := make([]types.Query, 0)
    for k, v := range r.URL.Query() {
        parser, exists := parsers[k]
        if !exists {
            log.Printf("Could not find query %s for data type %s\n", k, dataType)
            continue
        }
        query, err := parser(v)
        if err != nil {
            log.Printf("Error parsing query param %s for query %s: %v\n", v, k, err)
            continue
        }
        queries = append(queries, query)
    }

    data, err := s.db.GetByQueries(metaData, queries, ownerId)
    if err != nil {
        log.Println("Could not find data:", err)
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}

func (s *Server) handleGetByID(w http.ResponseWriter, r *http.Request) {
    ownerId, err := getOwnerIdFromRequestHeaders(r)
    if err != nil {
        log.Println("Could not get ownerId from headers", err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    vars := mux.Vars(r)
    dataType := vars["dataType"]
    metaData, exists := types.MetaDataMap[dataType]
    if !exists {
        log.Println("No metaData for specified data type:", dataType)
        http.Error(w, "Not Found", http.StatusBadRequest)
        return
    }

    id, err := strconv.ParseInt(vars["id"], 10, 64)
    if err != nil {
        log.Println("Could not parse int from id: ", vars["id"], err)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    data, err := s.db.Get(metaData, id, ownerId)
    if err != nil {
        log.Println("Could not find data:", err)
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}
