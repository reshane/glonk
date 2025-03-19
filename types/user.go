package types

import (
    "net/http"
    "encoding/json"
)

type User struct {
    ID int `json:"id"`
    Name string `json:"name"`
}

func (u User) Fields() []string {
    return []string{ "id", "name" }
}

func (u User) TableName() string {
    return "users"
}

func (u User) IntoRow() []any {
    return []any{ u.ID, u.Name }
}

func (u User) Validate() bool {
    return len(u.Name) > 0
}

func DecodeUserJson(r *http.Request) (DataType, error) {
    var user User
    err := json.NewDecoder(r.Body).Decode(&user)
    return user, err
}
