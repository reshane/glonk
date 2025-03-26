package types

import (
    "net/http"
    "reflect"
    "encoding/json"
)

// User data type
type User struct {
    ID int64 `json:"id" glonk:"id,owner_id"`
    Guid string `json:"guid" glonk:"guid"`
    Name string `json:"name" glonk:"name"`
    Email string `json:"email" glonk:"email"`
    Picture string `json:"picture" glonk:"picture"`
}

func (u User) TypeString() string {
    return userTypeString
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
type userMeta struct {}
var UserMeta userMeta = userMeta {}
var (
    UserQueries = Queries {}
    userTableName = "users"
    userFields = []string{ "id", "guid", "name", "email", "picture" }
    userTypeString = "user"
    userDecoder = DecodeUserJson
    userType = reflect.TypeOf(User{})
)

func (userMeta) GetType() reflect.Type {
    return userType
}

func (userMeta) TableName() string {
    return userTableName
}

func (userMeta) GetDecoder() Decoder {
    return DecodeUserJson
}

func (userMeta) GetQueries() Queries {
    return UserQueries
}

