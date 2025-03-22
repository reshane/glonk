package main

import (
    "os"
    "log"
    "strings"
    "context"

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
}
