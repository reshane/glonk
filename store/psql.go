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

var collectors map[string]func(pgx.CollectableRow) (types.DataType, error) = map[string]func(pgx.CollectableRow) (types.DataType, error) {
    types.UserMeta.TableName(): func (cr pgx.CollectableRow) (types.DataType, error) { res, err := pgx.RowToStructByName[types.User](cr); return res, err },
    types.NoteMeta.TableName(): func (cr pgx.CollectableRow) (types.DataType, error) { res, err := pgx.RowToStructByPos[types.Note](cr); return res, err },
}

func (s *PsqlStore) Get(metaData types.MetaData, id int64, ownerId int64) (types.DataType, error) {
    tableName := metaData.TableName()
    query := fmt.Sprintf("select %s from %s where %s=$1 and %s=$2", strings.Join(metaData.Fields(), " , "), tableName, metaData.IdField(), metaData.OwnerIdField())
    rows, err := s.conn.Query(context.Background(), query, id, ownerId)
    if err != nil {
        return nil, err
    }
    collector, exists := collectors[metaData.TableName()]
    if !exists {
        return nil, errors.New("No collector function for specified data type")
    }
    data, err := pgx.CollectOneRow(rows, collector)
    return data, err
}

func (s *PsqlStore) GetByQueries(metaData types.MetaData, queries []types.Query, ownerId int64) ([]types.DataType, error) {
    tableName := metaData.TableName()
    clauses := make([]string, 0)
    finalArgs := make([]any, 0)
    i := 1
    for _, query := range queries {
        // get the sql and argument map
        clause, args := query.Sql()
        for k, v := range args {
            // TODO: Add a check for collisions here
            // we can't have incompatible queries
            named := fmt.Sprintf("@%s", k)
            ordinal := fmt.Sprintf("$%d", i)
            clause = strings.Replace(clause, named, ordinal, -1)
            i += 1
            finalArgs = append(finalArgs, v)
        }
        clauses = append(clauses, "(" + clause + ")")
    }
    ownerIdClause := fmt.Sprintf("%s = $%d", metaData.OwnerIdField(), i)

    clauses = append(clauses, ownerIdClause)
    finalArgs = append(finalArgs, ownerId)

    query := fmt.Sprintf("select %s from %s where ", strings.Join(metaData.Fields(), " , "), tableName) + strings.Join(clauses, " and ")
    rows, err := s.conn.Query(context.Background(), query, finalArgs...)
    if err != nil {
        return nil, err
    }
    collector, exists := collectors[metaData.TableName()]
    if !exists {
        return nil, errors.New("No collector function for specified data type")
    }
    data, err := pgx.CollectRows(rows, collector)
    return data, err
}

func (s *PsqlStore) GetByGuid(metaData types.MetaData, guid string) (types.DataType, error) {
    tableName := metaData.TableName()
    query := fmt.Sprintf("select %s from %s where guid=$1", strings.Join(metaData.Fields(), " , "), tableName)
    rows, err := s.conn.Query(context.Background(), query, guid)
    if err != nil {
        return nil, err
    }
    collector, exists := collectors[metaData.TableName()]
    if !exists {
        return nil, errors.New("No collector function for specified data type")
    }
    data, err := pgx.CollectOneRow(rows, collector)
    return data, err
}

func (s *PsqlStore) Create(dataType types.DataType) (types.DataType, error) {
    // TODO: this pgx functionality should be used for a CreateMany function
    // change this to something like `insert into $1 ($2, $3...) values ($4, $5...) returning *;
    // and use the pgx.Query() function so that we can return the new value
    metaData, exists := types.MetaDataMap[dataType.TypeString()]
    if !exists {
        return nil, errors.New("No metadata found for specified dataType")
    }

    rows := [][]any{ dataType.IntoRow()[1:] }
    copyCount, err := s.conn.CopyFrom(
        context.Background(),
        pgx.Identifier{ metaData.TableName() },
        metaData.Fields()[1:],
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
    metaData, exists := types.MetaDataMap[dataType.TypeString()]
    if !exists {
        return nil, errors.New("No metadata found for specified dataType")
    }

    fieldMap := types.SparseUpdate(dataType)
    tableName := metaData.TableName()

    setStrings := make([]string, 0)
    values := make([]any, 0)
    var i int = 1
    for field, val := range fieldMap {
        if field != metaData.IdField() && field != metaData.OwnerIdField() {
            setStrings = append(setStrings, fmt.Sprintf("%s = $%d", field, i))
            values = append(values, val)
            i += 1
        }
    }
    values = append(values, dataType.GetId())
    values = append(values, dataType.GetOwnerId())
    fieldSetString := strings.Join(setStrings, ", ")

    query := fmt.Sprintf("update %s set %s where id = $%d and owner_id = $%d returning %s", tableName, fieldSetString, i, i + 1, strings.Join(metaData.Fields(), ","))
    rows, err := s.conn.Query(context.Background(), query, values...)
    if err != nil {
        return nil, err
    }
    collector, exists := collectors[metaData.TableName()]
    if !exists {
        log.Println("No collector function for specified table name:", metaData.TableName())
        return nil, errors.New("No collector function for specified data type")
    }
    data, err := pgx.CollectOneRow(rows, collector)
    return data, err
}

func (s *PsqlStore) Delete(metaData types.MetaData, id int64, owner_id int64) (types.DataType, error) {
    tableName := metaData.TableName()
    query := fmt.Sprintf("delete from %s where id=$1 and owner_id=$2 returning *", tableName)
    rows, err := s.conn.Query(context.Background(), query, id, owner_id)
    collector, exists := collectors[metaData.TableName()]
    if !exists {
        log.Println("No collector function for specified table name:", metaData.TableName())
        return nil, errors.New("No collector function for specified data type")
    }
    data, err := pgx.CollectOneRow(rows, collector)
    return data, err
}

