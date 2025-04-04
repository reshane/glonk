package store

import (
	"log"
    "reflect"
    "errors"
    "strings"
	"database/sql"
    "github.com/reshane/glonk/types"
)

type Store interface {
    // generic over DataType interface
    Get(types.MetaData, int64, int64) (types.DataType, error)
    GetByGuid(types.MetaData, string) (types.DataType, error)
    GetByQueries(types.MetaData, []types.Query, int64) ([]types.DataType, error)
    Create(types.DataType) (types.DataType, error)
    Update(types.DataType) (types.DataType, error)
    Delete(types.MetaData, int64, int64) (types.DataType, error)
}

// glonk internal reflection
func getGlonkName(field reflect.StructField) (string, error) {
    tagStr := field.Tag.Get("glonk")
    if tagStr == "" {
        return "", errors.New("No glonk value associated with field")
    }
    tags := strings.Split(tagStr, ",")
    return tags[0], nil
}

func getGlonkFieldIdx(a any, glonkField string) (int, error) {
    typ := reflect.TypeOf(a)
    val := reflect.ValueOf(a)
    for i := 0; i < typ.NumField(); i++ {
        glonkName, err := getGlonkName(typ.Field(i))
        if err != nil {
            return -1, errors.New("Could not get glonk name for " + val.Field(i).Type().Name())
        }
        if glonkName == glonkField {
            return i, nil
        }
    }
    return -1, errors.New("No glonk id for type " + typ.Name())
}

func fieldHasGlonkTag(field reflect.StructField, glonkTag string) bool {
    tagStr := field.Tag.Get(glonkTagStr)
    tags := strings.Split(tagStr, ",")
    for _, tag := range tags {
        if tag == glonkTag {
            return true
        }
    }
    return false
}


func getFromGlonkTag(a any, t string) (any, error) {
    val := reflect.ValueOf(a)
    typ := reflect.TypeOf(a)
    for i := 0; i < val.NumField(); i++ {
        typField := typ.Field(i)
        tag := typField.Tag.Get("glonk")
        if tag == "" {
            continue
        }
        valField := val.Field(i)
        tagVals := strings.Split(tag, ",")
        for _, tVal := range tagVals {
            if tVal == t && valField.CanInterface() {
                return valField.Interface(), nil
            }
        }
    }
    return nil, errors.New("No field with tag " + t + " on " + typ.Name())
}

// field specific glonk functions
var (
    glonkTagStr string = "glonk"
    glonkIdTag string = "id"
    glonkOwnerIdTag string = "owner_id"
    glonkAuthorIdTag string = "author_id"
)

func isId(field reflect.StructField) bool {
    return fieldHasGlonkTag(field, glonkIdTag)
}

func GetId(a any) int64 {
    idAny, err := getFromGlonkTag(a, glonkIdTag)
    if err != nil {
        return -1
    }
    id, ok := idAny.(int64)
    if !ok {
        return -1
    }
    return id
}

func getIdCol(typ reflect.Type) (string, error) {
    for i := 0; i < typ.NumField(); i++ {
        if isId(typ.Field(i)) {
            tagStr := typ.Field(i).Tag.Get(glonkTagStr)
            tags := strings.Split(tagStr, ",")
            return tags[0], nil
        }
    }
    return "", errors.New("No glonk id found for type " + typ.Name())
}

func isOwnerId(field reflect.StructField) bool {
    return fieldHasGlonkTag(field, glonkOwnerIdTag)
}

func GetOwnerId(a any) (int64, error) {
    ownerAny, err := getFromGlonkTag(a, glonkOwnerIdTag)
    if err != nil {
        return -1, err
    }
    ownerId, ok := ownerAny.(int64)
    if !ok {
        return -1, errors.New("owner_id type must be int64 and is set to " + reflect.TypeOf(ownerAny).Name() + " on " + reflect.TypeOf(a).Name())
    }
    return ownerId, nil
}

func getOwnerIdCol(typ reflect.Type) (string, error) {
    for i := 0; i < typ.NumField(); i++ {
        if isOwnerId(typ.Field(i)) {
            tagStr := typ.Field(i).Tag.Get(glonkTagStr)
            tags := strings.Split(tagStr, ",")
            return tags[0], nil
        }
    }
    return "", errors.New("No glonk owner_id found for type " + typ.Name())
}

func isAuthorId(field reflect.StructField) bool {
    return fieldHasGlonkTag(field, glonkAuthorIdTag)
}

func GetAuthorId(a any) (int64, error) {
    authorAny, err := getFromGlonkTag(a, glonkAuthorIdTag)
    if err != nil {
        return -1, err
    }
    authorId, ok := authorAny.(int64)
    if !ok {
        return -1, errors.New("author_id type must be int64 and is set to " + reflect.TypeOf(authorId).Name() + " on " + reflect.TypeOf(a).Name())
    }
    return authorId, nil
}

func getAuthorIdCol(typ reflect.Type) (string, error) {
    for i := 0; i < typ.NumField(); i++ {
        if isAuthorId(typ.Field(i)) {
            tagStr := typ.Field(i).Tag.Get(glonkTagStr)
            tags := strings.Split(tagStr, ",")
            return tags[0], nil
        }
    }
    return "", errors.New("No glonk author_id found for type " + typ.Name())
}

// glonk db functions
func intoSqlFields(typ reflect.Type) ([]string, error) {
    colNames := []string{glonkIdTag}
    fields := map[string]string{}
    for i := 0; i < typ.NumField(); i++ {
        if typ.Field(i).IsExported() {
            glonkName, err := getGlonkName(typ.Field(i))
            if err != nil {
                return []string{}, errors.New("Could not get column name for " + typ.Field(i).Name)
            }

            if fieldName, exists := fields[glonkName]; exists {
                return []string{}, errors.New("Fields " + fieldName + " and " + typ.Field(i).Name + " are tagged with the same glonk name: " + glonkName)
            }
            
            fields[glonkName] = typ.Field(i).Name
        }
    }
    if _, exists := fields[glonkIdTag]; exists {
        delete(fields, glonkIdTag)
    } else {
        return []string{}, errors.New("No glonk id found for type " + typ.Name())
    }
    if _, exists := fields[glonkOwnerIdTag]; exists {
        delete(fields, glonkOwnerIdTag)
        colNames = append(colNames, glonkOwnerIdTag)
    }
    for k, _ := range fields {
        colNames = append(colNames, k)
    }
    return colNames, nil
}

func intoRow(a any) ([]any, error) {
    row := []any{GetId(a)}
    val := reflect.ValueOf(a)
    typ := reflect.TypeOf(a)
    ownerIdCol, err := getOwnerIdCol(typ)
    if err == nil {
        idCol, err := getIdCol(typ)
        if err != nil {
            return []any{}, errors.New("No glonk id found for type " + typ.Name())
        }
        if idCol != ownerIdCol {
            ownerIdVal, err := GetOwnerId(a)
            if err != nil {
                return nil, err
            }
            row = append(row, ownerIdVal)
        }
    }
    for i := 0; i < val.NumField(); i++ {
        tagStr := typ.Field(i).Tag.Get("glonk")
        if tagStr != "" && (isOwnerId(typ.Field(i)) || isId(typ.Field(i))) {
            continue
        }
        if val.Field(i).CanInterface() {
            row = append(row, val.Field(i).Interface())
        }
    }
    return row, nil
}

func scanType(rows *sql.Rows, dt reflect.Type) ([]types.DataType, error) {
	fieldVals := make([]any, 0)
	for i := 0; i < dt.NumField(); i++ {
		reciever := reflect.New(dt.Field(i).Type)
		fieldVals = append(fieldVals, reciever.Elem().Interface())
	}

	fieldRefs := make([]any, 0)
	for i, _ := range fieldVals {
		fieldRefs = append(fieldRefs, &fieldVals[i])
	}

	data := make([]types.DataType, 0)
	for rows.Next() {
		targetDataPtr := reflect.New(dt)
		targetData := targetDataPtr.Elem()
		rows.Scan(fieldRefs...)
		for i, val := range fieldRefs {
			elem := reflect.ValueOf(val).Elem().Interface()
			switch t := elem.(type) {
			case int64:
				targetData.Field(i).Set(reflect.ValueOf(t))
			case float64:
				targetData.Field(i).Set(reflect.ValueOf(t))
			case string:
				targetData.Field(i).Set(reflect.ValueOf(t))
			default:
				log.Println("missing type", reflect.TypeOf(t))
			}
		}
		td, ok := targetData.Interface().(types.DataType)
		if ok {
			data = append(data, td)
		}
	}
	return data, nil
}

func sparseUpdate(dt types.DataType) (map[string]any, error) {
    fields, err := intoSqlFields(reflect.TypeOf(dt))
    if err != nil {
        return nil, err
    }

    vals, err := intoRow(dt)
    if err != nil {
        return nil, err
    }
    resultMap := make(map[string]any, 0)
    for i := 0; i < len(fields); i++ {
        if vals[i] != nil {
            resultMap[fields[i]] = vals[i]
        }
    }
    return resultMap, nil
}

type NoRows struct {}
func (NoRows) Error() string {
	return "No rows found"
}
