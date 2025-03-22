package types

import (
    "net/http"
)

var TypeStrings map[string]struct{} = map[string]struct{}{
    "user": struct{}{},
    "note": struct{}{},
}

var TypeStringToTableName map[string]string = map[string]string {
    "user": "users",
    "note": "notes",
}

var Decoders map[string]func(*http.Request) (DataType, error) = map[string]func(*http.Request) (DataType, error) {
    "user": DecodeUserJson,
    "note": DecodeNoteJson,
}

var QueryParsers map[string]map[string]func([]string) (Query, error) = map[string]map[string]func([]string) (Query, error) {
    "user": UserQueries,
    "note": NoteQueries,
}

var MetaData map[string]TypeMetaData = map[string]TypeMetaData {
    "note": NoteMeta,
    "user": UserMeta,
}

type DataType interface {
    IntoRow() []any
    Validate() bool
    TypeString() string
    GetId() int64
    GetOwnerId() int64
}

type TypeMetaData interface {
    TableName() string
    Fields() []string
    OwnerIdField() string
    IdField() string
}

type Query interface {
    Sql() (string, map[string]any)
}

func SparseUpdate(dt DataType) map[string]any {
    fields := MetaData[dt.TypeString()].Fields()
    vals := dt.IntoRow()
    resultMap := make(map[string]any, 0)
    for i := 0; i < len(fields); i++ {
        if vals[i] != nil {
            resultMap[fields[i]] = vals[i]
        }
    }
    return resultMap
}

