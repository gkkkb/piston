package authorization

import (
	"encoding/json"
	"net/http"
	"regexp"
)

// JWTToken is struct for JWT from aleppo
type JWTToken struct {
	Token                string           `json:"token"`
	Scopes               string           `json:"scopes"`
	RefreshToken         string           `json:"refresh_token"`
	ResourceOwner        JWTResourceOwner `json:"resource_owner"`
	ApplicationID        int              `json:"application_id"`
	ApplicationName      string           `json:"application_name"`
	ApplicationScopes    string           `json:"application_scopes"`
	ApplicationUID       string           `json:"application_uid"`
	ApplicationOwnerID   int              `json:"application_owner_id"`
	ApplicationOwnerType string           `json:"application_owner_type"`
}

// JWTResourceOwner hold owner information in aleppo JWT
type JWTResourceOwner struct {
	ID       uint        `json:"id"`
	Username string      `json:"username"`
	Name     string      `json:"name"`
	Email    string      `json:"email"`
	Gender   string      `json:"gender"`
	Phone    string      `json:"phone"`
	Role     string      `json:"role"`
	Agent    bool        `json:"agent"`
	Spammer  bool        `json:"spammer"`
	O2OAgent JWTO2OAgent `json:"o2o_agent"`
	Avatar   JWTAvatar   `json:"avatar"`
}

// JWTO2OAgent hold agent information in aleppo JWT
type JWTO2OAgent struct {
	ID       uint `json:"id"`
	Status   uint `json:"status"`
	JoinedAt uint `json:"joined_at"`
}

// JWTAvatar hold avatar information in aleppo JWT
type JWTAvatar struct {
	ID  uint   `json:"id"`
	URL string `json:"url"`
}

// JWTAuthorization is struct to convert HttpRequest to JWTToken struct
type JWTAuthorization struct {
	Token JWTToken
}

// NewJWTAuthorization returns a pointer of JWTAuthorization instance
func NewJWTAuthorization(r *http.Request) (*JWTAuthorization, error) {
	token, err := getToken(r)
	if err != nil {
		return nil, err
	}

	return &JWTAuthorization{Token: token}, nil
}

// IsScopeAuthorized is function to check token contain some scope
func (a *JWTAuthorization) IsScopeAuthorized(scope string) bool {
	match, _ := regexp.MatchString(scope, a.Token.Scopes)
	if !match {
		return false
	}

	return true
}

func getToken(r *http.Request) (JWTToken, error) {
	tokenStr := r.Header.Get("Authorization")

	token := JWTToken{}

	match, _ := regexp.MatchString("^Token ", tokenStr)
	if !match {
		return token, MissingTokenError{}
	}

	// Strip the `Token` part
	pToken := tokenStr[6:]
	raw, err := New().Decode(pToken)
	if err != nil {
		return token, InvalidTokenError{}
	}

	json.Unmarshal(raw, &token)

	return token, nil
}

// MissingTokenError is error struct for if Authorization header invalid
type MissingTokenError struct {
}

func (a MissingTokenError) Error() string {
	return "Token not presence"
}

// InvalidTokenError is error struct if JWT token invalid
type InvalidTokenError struct {
}

func (a InvalidTokenError) Error() string {
	return "Invalid Token"
}
