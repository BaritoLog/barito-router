package router

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var HMAC_JWT_SECRET = []byte("D0987373-7B6D-4DCC-AABC-2D11685E02B4")

type SSOClient struct {
	oAuth2Config *oauth2.Config
}

type Oauth2State struct {
	Redirect string `json:"redirect"`
}

type OAuthData struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
	HD            string `json:"hd"`
}

func NewSSOClient(clientID, clientSecret, redirectURL string) *SSOClient {
	return &SSOClient{
		oAuth2Config: &oauth2.Config{
			RedirectURL:  redirectURL,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (s SSOClient) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/auth/login":
		s.HandleLogin(w, r)
		return
	case "/auth/callback":
		s.HandleCallback(w, r)
		return
	default:
		// check auth, if its not logged in redirect to /auth/login
		tokenCookie, _ := r.Cookie("token")
		email, _ := s.GetEmailFromJWTToken(tokenCookie)
		if email == "" {
			http.Redirect(w, r, fmt.Sprintf("/auth/login?redirect=%s", r.URL.Path), http.StatusFound)
			return
		}

		fmt.Printf("Authenticated as %q to call: %q\n", email, r.URL.Path)
	}
}

func (s *SSOClient) GetEmailFromJWTToken(c *http.Cookie) (string, error) {
	if c == nil {
		return "", errors.New("No cookie found")
	}
	token, err := jwt.Parse(c.Value, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return HMAC_JWT_SECRET, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if email, ok := claims["email"]; ok {
			return email.(string), nil
		}
	}

	return "", errors.New("Not valid jwt token")
}

func (s *SSOClient) CreateJWTToken(email string) (string, error) {
	jwtClaims := jwt.MapClaims{
		"email": email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	tokenString, err := token.SignedString(HMAC_JWT_SECRET)
	if err != nil {
		fmt.Println("Failed to create JWT token", err)
		return "", err
	}
	return tokenString, nil
}

func (s *SSOClient) HandleCallback(w http.ResponseWriter, r *http.Request) {
	stateB64 := r.FormValue("state")
	stateStr, _ := base64.StdEncoding.DecodeString(stateB64)
	state := &Oauth2State{}
	json.Unmarshal([]byte(stateStr), &state)

	code := r.FormValue("code")
	data, err := s.getUserData(code)
	if err != nil {
		log.Errorf("Error when getting user data: %q", err.Error())
		w.Write([]byte("Failed to login, please ask the administrator"))
		return
	}

	jwtToken, err := s.CreateJWTToken(data.Email)
	if err != nil {
		w.Write([]byte("Failed to login, please ask the administrator"))
		return
	}
	cookie := http.Cookie{Name: "token", Value: jwtToken, Path: "/", Expires: time.Now().Add(24 * time.Hour)}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, state.Redirect, http.StatusFound)
}

func (s *SSOClient) getUserData(code string) (*OAuthData, error) {
	token, err := s.oAuth2Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var oAuthData OAuthData
	parse_err := json.Unmarshal(data, &oAuthData)
	if parse_err != nil {
		return nil, parse_err
	}
	return &oAuthData, nil
}

func (s *SSOClient) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state := Oauth2State{
		Redirect: r.FormValue("redirect"),
	}
	stateStr, _ := json.Marshal(state)
	stateB64 := base64.StdEncoding.EncodeToString([]byte(string(stateStr)))
	url := s.oAuth2Config.AuthCodeURL(string(stateB64))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
