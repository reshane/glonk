package types

import (
    "fmt"
    "strings"
    "strconv"
    "net/http"
    "encoding/json"
)

// Note data type
type Note struct {
    ID int64 `json:"id"`
    OwnerId int64 `json:"owner_id"`
    Contents string `json:"contents"`
}

var (
    NoteQueries = map[string]func([]string) (Query, error){
        "byOwnerId": ByOwnerIdFromQueryParam,
        "byContentContains": ByContentContainsFromQueryParam,
    }
    noteFields = []string{ "id", "owner_id", "contents" }
    noteTableName = "notes"
    noteTypeString = "note"
)

// DataType implementation for Note
func (n Note) Fields() []string {
    return noteFields
}

func (n Note) TableName() string {
    return noteTableName
}

func (n Note) IntoRow() []any {
    return []any{ n.ID, n.OwnerId, n.Contents }
}

func (n Note) TypeString() string {
    return noteTypeString
}

func (n Note) Id() int64 {
    return n.ID
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

// Query types
type ByOwnerId struct {
    ownerIds []int64
}

func ByOwnerIdFromQueryParam(queryParams []string) (Query, error) {
    ownerIds := make([]int64, 0)
    for _, queryParam := range queryParams {
        ownerIdStrings := strings.Split(queryParam, "|")
        for _, ownerIdString := range ownerIdStrings {
            ownerId, err := strconv.ParseInt(ownerIdString, 10, 64)
            if err != nil {
                return nil, err
            }
            ownerIds = append(ownerIds, ownerId)
        }
    }
    return &ByOwnerId { ownerIds }, nil
}

func (q *ByOwnerId) Sql() (string, map[string]any) {
    clauses := make([]string, 0)
    args := make(map[string]any)
    for i := 0; i < len(q.ownerIds); i++ {
        clauses = append(clauses, fmt.Sprintf("owner_id = @ownerId%d", i))
        args[fmt.Sprintf("ownerId%d", i)] = q.ownerIds[i]
    }
    return strings.Join(clauses, " or "), args
}

type ByContentContains struct {
    content string
}

func ByContentContainsFromQueryParam(queryParam []string) (Query, error) {
    if len(queryParam) > 1 || len(queryParam) == 0 {
        return nil, fmt.Errorf("ByContentContains only allows 1 parameter")
    }
    return &ByContentContains{ queryParam[0] }, nil
}

func (q *ByContentContains) Sql() (string, map[string]any) {
    likeExprVal := "%" + q.content + "%"
    args := map[string]any{ "contentContains": likeExprVal }
    clause := "contents like @contentContains"
    return clause, args
}

