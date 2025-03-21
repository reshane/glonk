package types

import (
    "net/http"
)

var TypeStrings map[string]struct{} = map[string]struct{}{
    "user": struct{}{},
}

var TypeStringToTableName map[string]string = map[string]string {
    "user": "users",
}

var Decoders map[string]func(*http.Request) (DataType, error) = map[string]func(*http.Request) (DataType, error) {
    "user": DecodeUserJson,
}

type DataType interface {
    IntoRow() []any
    TableName() string
    Fields() []string
    Validate() bool
    TypeString() string
    Id() int64
}

func SparseUpdate(dt DataType) map[string]any {
    fields := dt.Fields()
    vals := dt.IntoRow()
    resultMap := make(map[string]any, 0)
    for i := 0; i < len(fields); i++ {
        if vals[i] != nil {
            resultMap[fields[i]] = vals[i]
        }
    }
    return resultMap
}

