package types

import (
    "reflect"
    "net/http"
    "encoding/json"
)

// Post data type
type Post struct {
    ID int64 `json:"id" glonk:"id"`
    AuthorId int64 `json:"author_id" glonk:"author_id"`
    Contents string `json:"contents" glonk:"contents"`
}

func (p Post) TypeString() string {
    return postTypeString
}

func (p Post) Validate() bool {
    return len(p.Contents) > 0
}

// Decoders
func DecodePostJson(r *http.Request) (DataType, error) {
    var post Post 
    err := json.NewDecoder(r.Body).Decode(&post)
    return post, err
}

// Note metadata
type postMeta struct {}
var PostMeta postMeta = postMeta {}
var (
    PostQueries = Queries {
        "byAuthorId": { "author_id", ByIdFieldFromQueryParam },
        "byContentContains": { "contents", ByContainsFromQueryParam },
    }
    postFields = []string{ "id", "author_id", "contents" }
    postTableName = "posts"
    postTypeString = "post"
    postDecoder = DecodePostJson
    postType = reflect.TypeOf(Post{})
)

func (postMeta) GetType() reflect.Type {
    return postType
}

func (postMeta) TableName() string {
    return postTableName
}

func (postMeta) GetDecoder() Decoder {
    return postDecoder
}

func (postMeta) GetQueries() Queries {
    return PostQueries
}

