package store

import (
    "fmt"
    "log"
    "context"
    "os"
    "errors"
    "strings"

    "github.com/jackc/pgx/v5"

    "github.com/reshane/glonk/types"
)

type PsqlStore struct {
    conn *pgx.Conn
}

func NewPsqlStore() (*PsqlStore, error) {
    conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
    if err != nil {
        return nil, err
    }
    return &PsqlStore{ conn: conn }, nil
}

var collectors map[string]func(pgx.Rows) (types.DataType, error) = map[string]func(pgx.Rows) (types.DataType, error) {
    "user": func(rows pgx.Rows) (types.DataType, error) { return pgx.CollectOneRow(rows, pgx.RowToStructByPos[types.User]) },
}

func (s *PsqlStore) Get(dataType string, id int64) (types.DataType, error) {
    tableName, exists := types.TypeStringToTableName[dataType]
    if !exists {
        return nil, errors.New("No table name for specified data type")
    }
    query := fmt.Sprintf("select * from %s where id=$1", tableName)
    rows, err := s.conn.Query(context.Background(), query, id)
    if err != nil {
        return nil, err
    }
    collector, exists := collectors[dataType]
    if !exists {
        return nil, errors.New("No collector function for specified data type")
    }
    data, err := collector(rows)
    return data, err
}

func (s *PsqlStore) GetByGuid(dataType string, guid string) (types.DataType, error) {
    tableName, exists := types.TypeStringToTableName[dataType]
    if !exists {
        return nil, errors.New("No table name for specified data type")
    }
    query := fmt.Sprintf("select * from %s where guid=$1", tableName)
    rows, err := s.conn.Query(context.Background(), query, guid)
    if err != nil {
        return nil, err
    }
    collector, exists := collectors[dataType]
    if !exists {
        return nil, errors.New("No collector function for specified data type")
    }
    data, err := collector(rows)
    return data, err
}

func (s *PsqlStore) Create(dataType types.DataType) (types.DataType, error) {
    // TODO: this pgx functionality should be used for a CreateMany function
    // change this to something like `insert into $1 ($2, $3...) values ($4, $5...) returning *;
    // and use the pgx.Query() function so that we can return the new value
    rows := [][]any{ dataType.IntoRow()[1:] }
    copyCount, err := s.conn.CopyFrom(
        context.Background(),
        pgx.Identifier{ dataType.TableName() },
        dataType.Fields()[1:],
        pgx.CopyFromRows(rows),
    )
    if err != nil {
        return nil, err
    }
    if copyCount != 1 {
        return nil, errors.New("Could not create record")
    }
    return dataType, nil
}

func (s *PsqlStore) Update(dataType types.DataType) (types.DataType, error) {
    fieldMap := types.SparseUpdate(dataType)
    tableName := dataType.TableName()
    // turn fieldMap into `foo = @foo, bar = @bar`

    setStrings := make([]string, 0)
    values := make([]any, 0)
    var i int = 1
    for field, val := range fieldMap {
        if field != "id" {
            setStrings = append(setStrings, fmt.Sprintf("%s = $%d", field, i))
            values = append(values, val)
            i += 1
        }
    }
    values = append(values, dataType.Id())
    fieldSetString := strings.Join(setStrings, ", ")

    query := fmt.Sprintf("update %s set %s where id = $%d returning *", tableName, fieldSetString, i)
    rows, err := s.conn.Query(context.Background(), query, values...)
    if err != nil {
        return nil, err
    }
    collector, exists := collectors[dataType.TypeString()]
    if !exists {
        log.Println("No collector function for specified data type:", dataType.TypeString())
        return nil, errors.New("No collector function for specified data type")
    }
    data, err := collector(rows)
    return data, err
}

func (s *PsqlStore) Delete(dataType string, id int64) error {
    tableName, exists := types.TypeStringToTableName[dataType]
    if !exists {
        return errors.New("No table name for specified data type")
    }
    query := fmt.Sprintf("delete from %s where id=$1", tableName)
    commandTag, err := s.conn.Exec(context.Background(), query, id)
    if err != nil {
        return err
    }
    if commandTag.RowsAffected() != 1 {
        return errors.New("No row found to delete")
    }
    return nil
}

