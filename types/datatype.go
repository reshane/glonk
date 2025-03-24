package types

import (
    "strings"
    "encoding/json"
    "net/http"
)

var MetaDataMap map[string]MetaData = map[string]MetaData {
    "note": NoteMeta,
    "user": UserMeta,
}

// data struct interface
type DataType interface {
    IntoRow() []any
    Validate() bool
    TypeString() string
    GetId() int64
    GetOwnerId() int64
}

// data type metadata interface
type MetaData interface {
    json.Marshaler
    TableName() string
    Fields() []string
    OwnerIdField() string
    IdField() string
    GetDecoder() Decoder
    GetQueries() Queries
}
type Decoder = func(*http.Request) (DataType, error)
type Queries = map[string]func([]string) (Query, error)

func MarshalMetaDataJSON(md MetaData) ([]byte, error) {
    queryNames := make([]string, 0)
    for query, _ := range md.GetQueries() {
        queryNames = append(queryNames,`"` + query + `"`)
    }
    jsonString := `{"queries":[` + strings.Join(queryNames, ",") + `]}`
    return []byte(jsonString), nil
}

type Query interface {
    Sql() (string, map[string]any)
}

// turn data struct into a map from field names to values
func SparseUpdate(dt DataType) map[string]any {
    fields := MetaDataMap[dt.TypeString()].Fields()
    vals := dt.IntoRow()
    resultMap := make(map[string]any, 0)
    for i := 0; i < len(fields); i++ {
        if vals[i] != nil {
            resultMap[fields[i]] = vals[i]
        }
    }
    return resultMap
}

