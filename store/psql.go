package store

import (
    "fmt"
    "context"
    "os"
    "errors"

    "github.com/jackc/pgx/v5"

    "github.com/reshane/gorest/types"
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

func (s *PsqlStore) Get(dataType string, id int) (types.DataType, error) {
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

func (s *PsqlStore) Create(dataType types.DataType) (types.DataType, error) {
    rows := [][]any{ dataType.IntoRow() }
    copyCount, err := s.conn.CopyFrom(
        context.Background(),
        pgx.Identifier{ dataType.TableName() },
        dataType.Fields(),
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
    rows := [][]any{ dataType.IntoRow() }
    copyCount, err := s.conn.CopyFrom(
        context.Background(),
        pgx.Identifier{ dataType.TableName() },
        dataType.Fields(),
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

func (s *PsqlStore) Delete(dataType string, id int) error {
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

