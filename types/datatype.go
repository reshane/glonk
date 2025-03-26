package types

import (
    "strings"
    "reflect"
    "net/http"
)

type metaDataMap map[string]MetaData
var MetaDataMap metaDataMap = metaDataMap {
    "note": NoteMeta,
    "user": UserMeta,
    "post": PostMeta,
}

func (mdm metaDataMap) MarshalJSON() ([]byte, error) {
    jsonStrings := make([]string, 0)
    for dataType, md := range mdm {
        queryNames := make([]string, 0)
        for query, _ := range md.GetQueries() {
            queryNames = append(queryNames,`"` + query + `"`)
        }
        jsonStrings = append(jsonStrings, `"` + dataType + `":{"queries":[` + strings.Join(queryNames, ",") + `]}`)
    }
    return []byte(`{` + strings.Join(jsonStrings, ",") + `}`), nil
}

// data struct interface
type DataType interface {
    Validate() bool
    TypeString() string
}

// data type metadata interface
type MetaData interface {
    GetType() reflect.Type
    TableName() string
    Fields() []string
    GetDecoder() Decoder
    GetQueries() Queries
}
type Decoder = func(*http.Request) (DataType, error)
type QueryBuilder struct {
    Field string
    Parser func(string, []string) (Query, error)
}
type Queries = map[string]QueryBuilder

type Query interface {
    Sql() (string, map[string]any)
}


