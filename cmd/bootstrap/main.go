package main

import (
    "os"
    "flag"
    "log"
    "strings"
    "context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
    "github.com/jackc/pgx/v5"
)

func ReadToString(path string) (string, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        log.Printf("Error: Could not read file %s: %s", path, err)
        return "", err
    }
    return string(content), nil
}

func main() {
    whichDb := flag.String("storage", "sqlite3", "The data storeage to use - psql: Postgres, sqlite3: Sqlite3 (default)")
    flag.Parse()

	if *whichDb == "psql" {
        conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
        if err != nil {
            log.Fatal(err)
        }
        contents, err := ReadToString("./db/scripts/bootstrap.sql")
        if err != nil {
            log.Fatal(err)
        }
        migrations := strings.Split(contents, "-- @COMMAND")
        for _, migration := range migrations {
            if len(migration) == 0 {
                continue
            }
            commandTag, err := conn.Exec(context.Background(), migration)
            if err != nil {
                log.Fatal(err)
            }
            log.Println(migration, commandTag)
        }
    } else {
        conn, err := sql.Open("sqlite3", "./test.db")
        if err != nil {
            log.Fatal(err);
        }
        contents, err := ReadToString("./db/scripts/bootstrap_sqlite.sql")
        if err != nil {
            log.Fatal(err)
        }
        migrations := strings.Split(contents, "-- @COMMAND")
        for _, migration := range migrations {
            log.Println(migration);
            if len(migration) == 0 {
                continue
            }
            _, err := conn.Exec(migration)
            if err != nil {
                log.Fatal(err)
            }
        }
    }
}
