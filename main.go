package main

import (
    "log"
    "flag"

    "github.com/reshane/gorest/api"
    "github.com/reshane/gorest/store"
)

func main() {
    listenAddr := flag.String("listenaddr", ":8080", "The server address")
    flag.Parse()

    db, err := store.NewPsqlStore()
    if err != nil {
        log.Fatalf("Could not create db connection: %v", err)
    }

    server := api.NewServer(*listenAddr, db)
    log.Println("Server running on port: ", *listenAddr)
    log.Fatal(server.Start())
}
