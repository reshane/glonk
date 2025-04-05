package main

import (
    "log"
    "flag"

    "github.com/reshane/glonk/api"
    "github.com/reshane/glonk/store"
)

func getDb(which string) (store.Store, error) {
	if which == "psql" {
		return store.NewPsqlStore()
	}
	return store.NewSqliteStore()
}

func main() {
	listenAddr := flag.String("listenaddr", ":8080", "The server address (default :8080)")
	whichDb := flag.String("storage", "sqlite3", "The data storeage to use - psql: Postgres, sqlite3: Sqlite3 (default)")
    flag.Parse()

	db, err := getDb(*whichDb)
    if err != nil {
        log.Fatalf("Could not create db connection: %v", err)
    }

    server := api.NewServer(*listenAddr, db)
    log.Println("Server running on port: ", *listenAddr)
    log.Fatal(server.Start())
}
