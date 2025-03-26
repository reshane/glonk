package store

import (
    "fmt"
    "log"
    "context"
    "os"
    "errors"
    "strings"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/jackc/pgx/v5"

    "github.com/reshane/glonk/types"
)

type PsqlStore struct {
    conn *pgxpool.Pool
}

func NewPsqlStore() (*PsqlStore, error) {
    conn, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
    if err != nil {
        return nil, err
    }
    return &PsqlStore{ conn: conn }, nil
}

var collectors map[string]func(pgx.CollectableRow) (types.DataType, error) = map[string]func(pgx.CollectableRow) (types.DataType, error) {
    types.UserMeta.TableName(): func (cr pgx.CollectableRow) (types.DataType, error) { res, err := pgx.RowToStructByName[types.User](cr); return res, err },
    types.NoteMeta.TableName(): func (cr pgx.CollectableRow) (types.DataType, error) { res, err := pgx.RowToStructByName[types.Note](cr); return res, err },
    types.PostMeta.TableName(): func (cr pgx.CollectableRow) (types.DataType, error) { res, err := pgx.RowToStructByName[types.Post](cr); return res, err },
}

func (s *PsqlStore) Get(metaData types.MetaData, id int64, ownerId int64) (types.DataType, error) {
    dataType := metaData.GetType()
    tableName := metaData.TableName()

    finalArgs := []any{id}
    fields, err := intoSqlFields(dataType)
    if err != nil {
        log.Println("Could not retreive sql fields for ", dataType)
        return nil, err
    }
    clauses := "id = $1"
    ownerIdCol, err := getOwnerIdCol(dataType)
    if err == nil {
        clauses += " and " + ownerIdCol + " = $2"
        finalArgs = append(finalArgs, ownerId)
    }
    authorIdCol, err := getAuthorIdCol(dataType)
    if err == nil {
        clauses += " and " + authorIdCol + " = $2"
        finalArgs = append(finalArgs, ownerId)
    }

    query := fmt.Sprintf("select %s from %s where %s", strings.Join(fields, ","), tableName, clauses)
    rows, err := s.conn.Query(context.Background(), query, finalArgs...)
    if err != nil {
        return nil, err
    }
    collector, exists := collectors[metaData.TableName()]
    if !exists {
        return nil, errors.New("No collector function for specified data type")
    }
    return pgx.CollectOneRow(rows, collector)
}

func (s *PsqlStore) GetByQueries(metaData types.MetaData, queries []types.Query, ownerId int64) ([]types.DataType, error) {
    dataType := metaData.GetType()
    tableName := metaData.TableName()

    fields, err := intoSqlFields(dataType)
    if err != nil {
        log.Println("Could not retreive sql fields for ", dataType)
        return nil, err
    }
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

    ownerIdCol, err := getOwnerIdCol(metaData.GetType())
    if err == nil {
        ownerIdClause := fmt.Sprintf("%s = $%d", ownerIdCol, i)
        clauses = append(clauses, ownerIdClause)
        finalArgs = append(finalArgs, ownerId)
        i += 1
    }
    authorIdCol, err := getAuthorIdCol(dataType)
    if err == nil {
        authorIdClause := fmt.Sprintf("%s = $%d", authorIdCol, i)
        clauses = append(clauses, authorIdClause)
        finalArgs = append(finalArgs, ownerId)
    }


    query := fmt.Sprintf("select %s from %s", strings.Join(fields, ","), tableName)
    if len(clauses) > 0 {
        query += " where " + strings.Join(clauses, " and ")
    }
    rows, err := s.conn.Query(context.Background(), query, finalArgs...)
    if err != nil {
        return nil, err
    }
    collector, exists := collectors[metaData.TableName()]
    if !exists {
        return nil, errors.New("No collector function for specified data type")
    }
    return pgx.CollectRows(rows, collector)
}

func (s *PsqlStore) GetByGuid(metaData types.MetaData, guid string) (types.DataType, error) {
    dataType := metaData.GetType()
    tableName := metaData.TableName()
    fields, err := intoSqlFields(dataType)
    if err != nil {
        log.Println("Could not retreive sql fields for ", dataType)
        return nil, err
    }
    query := fmt.Sprintf("select %s from %s where guid=$1", strings.Join(fields, ","), tableName)
    rows, err := s.conn.Query(context.Background(), query, guid)
    if err != nil {
        return nil, err
    }
    collector, exists := collectors[metaData.TableName()]
    if !exists {
        return nil, errors.New("No collector function for specified data type")
    }
    return pgx.CollectOneRow(rows, collector)
}

func (s *PsqlStore) Create(data types.DataType) (types.DataType, error) {
    metaData, exists := types.MetaDataMap[data.TypeString()]
    if !exists {
        return nil, errors.New("No metadata found for specified dataType")
    }
    dataType := metaData.GetType()

    rowVals, err := intoRow(data)
    if err != nil {
        return nil, err
    }
    values := rowVals[1:]
    allFields, err := intoSqlFields(dataType)
    if err != nil {
        log.Println("Could not retreive sql fields for ", dataType)
        return nil, err
    }
    fields := allFields[1:]
    placeholders := make([]string, 0)
    for i, _ := range fields {
        placeholders = append(placeholders, fmt.Sprintf("$%d", i + 1))
    }
    fieldString := strings.Join(fields, ",")
    placeholderString := strings.Join(placeholders, ",")

    query := fmt.Sprintf("insert into %s (%s) values (%s) returning *", metaData.TableName(), fieldString, placeholderString)
    rows, err := s.conn.Query(context.Background(), query, values...)
    if err != nil {
        return nil, err
    }
    collector, exists := collectors[metaData.TableName()]
    if !exists {
        log.Println("No collector function for specified table name:", metaData.TableName())
        return nil, errors.New("No collector function for specified data type")
    }
    return pgx.CollectOneRow(rows, collector)
    // TODO: this pgx functionality should be used for a CreateMany function
    // change this to something like `insert into $1 ($2, $3...) values ($4, $5...) returning *;
    // and use the pgx.Query() function so that we can return the new value
    /*rows := [][]any{ dataType.IntoRow()[1:] }
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
    return dataType, nil*/
}

func (s *PsqlStore) Update(data types.DataType) (types.DataType, error) {
    metaData, exists := types.MetaDataMap[data.TypeString()]
    if !exists {
        return nil, errors.New("No metadata found for specified dataType")
    }
    dataType := metaData.GetType()
    fields, err := intoSqlFields(dataType)
    if err != nil {
        log.Println("Could not retreive sql fields for ", dataType)
        return nil, err
    }

    fieldMap, err := sparseUpdate(data)
    if err != nil {
        return nil, err
    }
    tableName := metaData.TableName()

    setStrings := make([]string, 0)
    values := make([]any, 0)
    var i int = 1
    for field, val := range fieldMap {
        if field != glonkIdTag && field != glonkOwnerIdTag {
            setStrings = append(setStrings, fmt.Sprintf("%s = $%d", field, i))
            values = append(values, val)
            i += 1
        }
    }
    values = append(values, GetId(data))
    ownerIdValue, err := GetOwnerId(data)
    authorOwnerField := "owner_id"
    if err != nil {
        ownerIdValue, err = GetAuthorId(data)
        if err != nil {
            return nil, err
        }
        authorOwnerField = "author_id"
    }
    values = append(values, ownerIdValue)
    fieldSetString := strings.Join(setStrings, ", ")

    query := fmt.Sprintf("update %s set %s where id = $%d and %s = $%d returning %s", tableName, fieldSetString, i, authorOwnerField, i + 1, strings.Join(fields, ","))
    rows, err := s.conn.Query(context.Background(), query, values...)
    if err != nil {
        return nil, err
    }
    collector, exists := collectors[metaData.TableName()]
    if !exists {
        log.Println("No collector function for specified table name:", metaData.TableName())
        return nil, errors.New("No collector function for specified data type")
    }
    return pgx.CollectOneRow(rows, collector)
}

func (s *PsqlStore) Delete(metaData types.MetaData, id int64, owner_id int64) (types.DataType, error) {
    tableName := metaData.TableName()
    query := fmt.Sprintf("delete from %s where id=$1 and owner_id=$2 returning *", tableName)
    rows, err := s.conn.Query(context.Background(), query, id, owner_id)
    if err != nil {
        return nil, err
    }
    collector, exists := collectors[metaData.TableName()]
    if !exists {
        log.Println("No collector function for specified table name:", metaData.TableName())
        return nil, errors.New("No collector function for specified data type")
    }
    return pgx.CollectOneRow(rows, collector)
}

