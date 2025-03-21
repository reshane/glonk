package types

import (
    "net/http"
    "encoding/json"
)

type Note struct {
    ID int64 `json:"id"`
    OwnerId int64 `json:"owner_id"`
    Contents string `json:"contents"`
}

func (n Note) Fields() []string {
    return []string{ "id", "owner_id", "contents" }
}

func (n Note) TableName() string {
    return "notes"
}

func (n Note) IntoRow() []any {
    return []any{ n.ID, n.OwnerId, n.Contents }
}

func (n Note) TypeString() string {
    return "note"
}

func (n Note) Id() int64 {
    return n.ID
}

func (n Note) Validate() bool {
    return len(n.Contents) > 0
}

func DecodeNoteJson(r *http.Request) (DataType, error) {
    var note Note
    err := json.NewDecoder(r.Body).Decode(&note)
    return note, err
}
