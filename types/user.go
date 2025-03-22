package types

import (
    "net/http"
    "encoding/json"
)

type User struct {
    ID int64 `json:"id"`
    Guid string `json:"guid"`
    Email string `json"email"`
    Name string `json:"name"`
    Picture string `json"picture"`
}

var (
    UserQueries = map[string]func([]string) (Query, error){}
    userTableName = "users"
    userFields = []string{ "id", "guid", "email", "name", "picture" }
    userTypeString = "user"
)

func (u User) TableName() string {
    return userTableName
}

func (u User) Fields() []string {
    return userFields
}

func (u User) IntoRow() []any {
    return []any{ u.ID, u.Guid, u.Email, u.Name, u.Picture }
}

func (u User) TypeString() string {
    return userTypeString
}

func (u User) Id() int64 {
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

