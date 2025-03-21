package types

import (
    "net/http"
    "encoding/json"
)

type User struct {
    ID int64 `json:"id"`
    Guid string `json:"guid"`
    Name string `json:"name"`
}

func (u User) Fields() []string {
    return []string{ "id", "guid", "name" }
}

func (u User) TableName() string {
    return "users"
}

func (u User) IntoRow() []any {
    return []any{ u.ID, u.Guid, u.Name }
}

func (u User) TypeString() string {
    return "user"
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
