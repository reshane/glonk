package store

import (
	"log"
	"fmt"
	"errors"
	"strings"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"github.com/reshane/glonk/types"
)

type SqliteStore struct {
	conn *sql.DB
}

func NewSqliteStore() (*SqliteStore, error) {
    conn, err := sql.Open("sqlite3", "./test.db")
    if err != nil {
        return nil, err
    }
	return &SqliteStore{ conn: conn }, nil
}


func (s *SqliteStore) Get(metaData types.MetaData, id int64, owner_id int64) (types.DataType, error) {
	dataType := metaData.GetType()
	tableName := metaData.TableName()
    fields, err := intoSqlFields(dataType)
    if err != nil {
        log.Println("Could not retreive sql fields for ", dataType)
        return nil, err
    }

	query := fmt.Sprintf("SELECT %s FROM %s where id = (?)", strings.Join(fields, ","), tableName)
	vals := []any{id}
    ownerIdCol, err := getOwnerIdCol(metaData.GetType())
    if err == nil {
		query += fmt.Sprintf(" and %s = (?)", ownerIdCol)
		vals = append(vals, owner_id)
    }

	statement, err := s.conn.Prepare(query)

	rows, err := statement.Query(vals...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	data, err := scanType(rows, dataType)
	if err != nil {
		return nil, err
	}
	if len(data) > 1 {
		return nil, errors.New(fmt.Sprintf("Multiple entries found for id: %d, owner_id: %d - %v", id, owner_id, data))
	}
	if len(data) == 0 {
		return nil, NoRows{}
	}
	return data[0], nil
}

func (s *SqliteStore) GetByGuid(metaData types.MetaData, guid string) (types.DataType, error) {
	dataType := metaData.GetType()
	tableName := metaData.TableName()
    fields, err := intoSqlFields(dataType)
    if err != nil {
        log.Println("Could not retreive sql fields for ", dataType)
        return nil, err
    }
	query := fmt.Sprintf("SELECT %s FROM %s where guid = (?)", strings.Join(fields, ","), tableName)
	statement, err := s.conn.Prepare(query)
	rows, err := statement.Query(guid)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	data, err := scanType(rows, dataType)
	if err != nil {
		return nil, err
	}
	if len(data) > 1 {
		return nil, errors.New(fmt.Sprintf("Multiple entries found for guid: %s, %v", guid, data))
	}
	if len(data) == 0 {
		return nil, NoRows{}
	}
	return data[0], nil
}

func (s *SqliteStore) GetByQueries(metaData types.MetaData, queries []types.Query, ownerId int64) ([]types.DataType, error) {
    dataType := metaData.GetType()
    tableName := metaData.TableName()

    fields, err := intoSqlFields(dataType)
    if err != nil {
        log.Println("Could not retreive sql fields for ", dataType)
        return nil, err
    }
    clauses := make([]string, 0)
    finalArgs := make([]any, 0)
    for _, query := range queries {
        clause, args := query.Sql()
        for k, v := range args {
            named := fmt.Sprintf("@%s", k)
            clause = strings.Replace(clause, named, "?", -1)
            finalArgs = append(finalArgs, v)
        }
        clauses = append(clauses, "(" + clause + ")")
    }

    ownerIdCol, err := getOwnerIdCol(metaData.GetType())
    if err == nil {
        ownerIdClause := fmt.Sprintf("%s = ?", ownerIdCol)
        clauses = append(clauses, ownerIdClause)
        finalArgs = append(finalArgs, ownerId)
    }

    query := fmt.Sprintf("select %s from %s", strings.Join(fields, ","), tableName)
    if len(clauses) > 0 {
        query += " where " + strings.Join(clauses, " and ")
    }
	statement, err := s.conn.Prepare(query)

	rows, err := statement.Query(finalArgs...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	data, err := scanType(rows, dataType)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *SqliteStore) Create(data types.DataType) (types.DataType, error) {
    metaData, exists := types.MetaDataMap[data.TypeString()]
    if !exists {
        return nil, errors.New("No metadata found for specified dataType")
    }
    dataType := metaData.GetType()
    tableName := metaData.TableName()

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
    for _, _ = range fields {
        placeholders = append(placeholders, "?")
    }
    fieldString := strings.Join(fields, ",")
    placeholderString := strings.Join(placeholders, ",")

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s", tableName, fieldString, placeholderString, strings.Join(allFields, ","))
	statement, err := s.conn.Prepare(query)
	if err != nil {
		return nil, err
	}

	rows, err := statement.Query(values...)
	if err != nil {
		return nil, err
	}

	created, err := scanType(rows, dataType)
	if err != nil {
		return nil, err
	}
	if len(created) != 1 {
		return nil, errors.New(fmt.Sprintf("Multiple (%d) entries created for %v", len(created), created))
	}
	return created[0], nil
}

func (s *SqliteStore) Update(data types.DataType) (types.DataType, error) {
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
	statement, err := s.conn.Prepare(query)
	rows, err := statement.Query(values...)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	updated, err := scanType(rows, dataType)
	if err != nil {
		return nil, err
	}
	if len(updated) != 1 {
		return nil, errors.New(fmt.Sprintf("Multiple (%d) entries updated for id: %d, owner_id: %d", len(updated), GetId(data), ownerIdValue))
	}
	return updated[0], nil
}

func (s *SqliteStore) Delete(metaData types.MetaData, id int64, owner_id int64) (types.DataType, error) {
	// TODO: this should probably query first and then delete
	// if multiple entries are found, we should report the error rather than
	// reporting multiple entries were deleted
	dataType := metaData.GetType()
	tableName := metaData.TableName()
    fields, err := intoSqlFields(dataType)
    if err != nil {
        log.Println("Could not retreive sql fields for ", dataType)
        return nil, err
    }
	query := fmt.Sprintf("DELETE FROM %s where id = (?) and owner_id = (?) returning %s", tableName, strings.Join(fields, ","))
	statement, err := s.conn.Prepare(query)
	rows, err := statement.Query(id, owner_id)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	data, err := scanType(rows, dataType)
	if err != nil {
		return nil, err
	}
	if len(data) != 1 {
		return nil, errors.New(fmt.Sprintf("Multiple (%d) entries deleted for id: %d, owner_id: %d", len(data), id, owner_id))
	}
	return data[0], nil
}
