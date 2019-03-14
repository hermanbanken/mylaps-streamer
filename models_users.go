package main

import "time"

// LoginJSON is the output of the login
type LoginJSON struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	UserID      string `json:"userId"`
	IssuedAt    string `json:".issued"`
	ExpiresAt   string `json:".expires"`
}

// SavedUser is saved in the db
type SavedUser struct {
	Name         string
	UserID       string
	AccessToken  string
	MadePublicAt time.Time
}

// Example: {
// 	"access_token": "flurp",
// 	"token_type": "bearer",
// 	"expires_in": 15767999,
// 	"userId": "MYLAPS-GA-flurp",
// 	".issued": "Tue, 29 Jan 2019 19:51:58 GMT",
// 	".expires": "Wed, 31 Jul 2019 07:51:58 GMT"
// }

// ProfileJSON is the basic short profile info
type ProfileJSON struct {
	UserID        string `json:"userId"`
	GivenName     string `json:"givenName"`
	SurName       string `json:"surName"`
	NickName      string `json:"nickName"`
	Email         string `json:"email"`
	AccountStatus string `json:"accountStatus"`
}

// ClaimsJSON returned by /auth/claims
type ClaimsJSON struct {
	Claims []struct {
		Subject   string
		Type      string
		Value     string
		ValueType string
	}
}

// Example:
// {"claims":[
// 	{"subject":"me","type":"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier","value":"MYLAPS-GA-yadayada","valueType":"http://www.w3.org/2001/XMLSchema#string"},
// 	{"subject":"me","type":"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress","value":"me@provider.com","valueType":"http://www.w3.org/2001/XMLSchema#string"},
// 	{"subject":"me","type":"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name","value":"me","valueType":"http://www.w3.org/2001/XMLSchema#string"},
// 	{"subject":"me","type":"http://schemas.microsoft.com/ws/2008/06/identity/claims/role","value":"user","valueType":"http://www.w3.org/2001/XMLSchema#string"}
// ]}
