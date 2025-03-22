package types

import (
    "net/http"
    "encoding/json"
)

type User struct {
    ID int64 `json:"id"`
    Guid string `json:"guid"`
    Name string `json:"name"`
    Email string `json:"email"`
    Picture string `json:"picture"`
}

var (
    UserQueries = map[string]func([]string) (Query, error){}
    userTableName = "users"
    userFields = []string{ "id", "guid", "name", "email", "picture" }
    userTypeString = "user"
)

// User metadata
type userMeta struct{}
var UserMeta userMeta

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

func DecodeUserJson(r *http.Request) (DataType, error) {
    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    return user, err
}

