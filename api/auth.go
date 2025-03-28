package api

import (
    "net/http"
    "log"
    "fmt"
    "os"
    "time"
    "strconv"
    "crypto/rand"
    "context"
    "encoding/base64"
    "encoding/json"
    "errors"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"

    "github.com/jackc/pgx/v5"

    "github.com/reshane/glonk/types"
)

// session struct
type session struct {
    userId string
    ownerId int64
    expiry time.Time
}

// google user response object
type UserInfo struct {
    Id string `json"id"`
    Email string `json"email"`
    VerifiedEmail bool `json"verified_email"`
    Name string `json"name"`
    GivenName string `json"given_name"`
    FamilyName string `json"family_name"`
    Picture string `json"picture"`
    Locale string `json"locale"`
}

var (
    // sessions map
    sessions = map[string]session{}

    // google oauth config
    cfg = &oauth2.Config{
        RedirectURL: "http://localhost:8080/auth/google/callback",
        ClientID: os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
        ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
        Scopes: []string{"email", "profile"},
        Endpoint: google.Endpoint,
    }
)

// google user info endpoint
const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

// authorization middleware
func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        sessionId, err := r.Cookie("session_id")
        if err != nil {
            if err == http.ErrNoCookie {
                http.Error(w, "Not Authorized", http.StatusUnauthorized)
                return
            }
            http.Error(w, "Bad Request", http.StatusBadRequest)
            return
        }
        session, exists := sessions[sessionId.Value]
        if exists && session.expiry.After(time.Now()) {
            r.Header.Set("OwnerId", strconv.FormatInt(session.ownerId, 10))
            endpoint(w, r)
            return
        }
        http.Error(w, "Not Authorized", http.StatusUnauthorized)
        return
    })
}

// loging endpoint & callback
func (s *Server) googleLogin(w http.ResponseWriter, r *http.Request) {
    oauthState := generateStateCookie(w)
    u := cfg.AuthCodeURL(oauthState)
    http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func (s *Server) googleCallback(w http.ResponseWriter, r *http.Request) {
    oauthState, err := r.Cookie("oauthstate")
    if err != nil {
        if err == http.ErrNoCookie {
            http.Error(w, "Not Authorized", http.StatusUnauthorized)
            return
        }
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    if r.FormValue("state") != oauthState.Value {
        log.Println("Invalid oauth google state")
        http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
        return
    }

    userInfo, err := getUserDataFromGoogle(r.FormValue("code"))
    if err != nil {
        log.Println(err.Error())
        http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
        return
    }

    // redirect to user endpoint
    retrievedUser, err := s.retreiveOrCreateUser(userInfo)
    if err != nil {
        log.Println(err.Error())
        http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
        return
    }

    var expiration = time.Now().Add(20 * time.Minute)
    b := make([]byte, 16)
    rand.Read(b)
    sessionId := base64.URLEncoding.EncodeToString(b)
    cookie := http.Cookie{
        Name: "session_id",
        Value: sessionId,
        Path: "/",
        HttpOnly: true,
        Expires: expiration,
    }
    http.SetCookie(w, &cookie)
    sessions[sessionId] = session {
        userId: retrievedUser.Guid,
        ownerId: retrievedUser.ID,
        expiry: expiration,
    }

    http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s *Server) retreiveOrCreateUser(userInfo *UserInfo) (*types.User, error) {
    guid := userInfo.Id
    user, err := s.db.GetByGuid(types.UserMeta, "google/" + guid)
    if err != nil {
        if !errors.Is(err, pgx.ErrNoRows) {
            return nil, err
        }

        newUser := types.User{
            Guid: "google/" + guid,
            Name: userInfo.Name,
            Email: userInfo.Email,
            Picture: userInfo.Picture,
        }
        log.Println("Creating new user", newUser)
        user, err = s.db.Create(newUser)
        if err != nil {
            log.Println("Error creating new user:", err.Error())
            return nil, err
        }
    }
    retreivedUser := user.(types.User)
    return &retreivedUser, nil
}

func generateStateCookie(w http.ResponseWriter) string {
    var expiration = time.Now().Add(20 * time.Minute)

    b := make([]byte, 16)
    rand.Read(b)
    state := base64.URLEncoding.EncodeToString(b)
    cookie := http.Cookie{ Name: "oauthstate", Value: state, Expires: expiration }
    http.SetCookie(w, &cookie)

    return state
}

func getUserDataFromGoogle(code string) (*UserInfo, error) {
    token, err := cfg.Exchange(context.Background(), code)
    if err != nil {
        return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
    }
    response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
    if err != nil {
        return nil, fmt.Errorf("failed getting user info: %s", err.Error())
    }
    defer response.Body.Close()
    var userInfo UserInfo
    if err = json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
        return nil, fmt.Errorf("failed to read response: %s", err.Error())
    }
    return &userInfo, nil
}

