package types

import (
    "net/http"
    "encoding/json"
)

// User data type
type User struct {
    ID int64 `json:"id"`
    Guid string `json:"guid"`
    Name string `json:"name"`
    Email string `json:"email"`
    Picture string `json:"picture"`
}

func (u User) IntoRow() []any {
    return []any{ u.ID, u.Guid, u.Name, u.Email, u.Picture }
}

func (u User) TypeString() string {
    return userTypeString
}

func (u User) GetId() int64 {
    return u.ID
}

func (u User) GetOwnerId() int64 {
    return u.ID
}

func (u User) Validate() bool {
    return len(u.Name) > 0 && len(u.Guid) > 0
}

// Decoders
func DecodeUserJson(r *http.Request) (DataType, error) {
    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    return user, err
}


// User metadata
type userMeta struct{}
var UserMeta userMeta
var (
    UserQueries = map[string]func([]string) (Query, error){}
    userTableName = "users"
    userFields = []string{ "id", "guid", "name", "email", "picture" }
    userTypeString = "user"
    userDecoder = DecodeUserJson
)

func (userMeta) OwnerIdField() string {
    return "id"
}

func (userMeta) IdField() string {
    return "id"
}

func (userMeta) TableName() string {
    return userTableName
}

func (userMeta) Fields() []string {
    return userFields
}

func (userMeta) GetDecoder() Decoder {
    return DecodeUserJson
}

func (userMeta) GetQueries() Queries {
    return UserQueries
}


