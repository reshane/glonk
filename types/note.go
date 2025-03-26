package types

import (
    "reflect"
    "net/http"
    "encoding/json"
)

// Note data type
type Note struct {
    ID int64 `json:"id" glonk:"id"`
    OwnerId int64 `json:"owner_id" glonk:"owner_id"`
    Contents string `json:"contents" glonk:"contents"`
}

func (n Note) IntoRow() []any {
    return []any{ n.ID, n.OwnerId, n.Contents }
}

func (n Note) TypeString() string {
    return noteTypeString
}

func (n Note) Validate() bool {
    return len(n.Contents) > 0
}

// Decoders
func DecodeNoteJson(r *http.Request) (DataType, error) {
    var note Note
    err := json.NewDecoder(r.Body).Decode(&note)
    return note, err
}

// Note metadata
type noteMeta struct {}
var NoteMeta noteMeta = noteMeta {}
var (
    NoteQueries = Queries {
        "byOwnerId": { "owner_id", ByIdFieldFromQueryParam },
        "byContentContains": { "contents", ByContainsFromQueryParam },
    }
    noteFields = []string{ "id", "owner_id", "contents" }
    noteTableName = "notes"
    noteTypeString = "note"
    noteDecoder = DecodeNoteJson
    noteType = reflect.TypeOf(Note{})
)

func (noteMeta) GetType() reflect.Type {
    return noteType
}

func (noteMeta) TableName() string {
    return noteTableName
}

func (noteMeta) GetDecoder() Decoder {
    return noteDecoder
}

func (noteMeta) GetQueries() Queries {
    return NoteQueries
}

