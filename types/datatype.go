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
}

