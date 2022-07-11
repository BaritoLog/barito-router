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

const (
	FAILED_LOGIN_ERROR_MESSAGE = "Failed to login, please ask the administrator"
	FORM_CODE_KEY              = "code"
	FORM_REDIRECT_KEY          = "redirect"
	FORM_STATE_KEY             = "state"
	GOOGLE_API_URL             = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	JWT_COOKIE_EMAIL_KEY       = "email"
	JWT_COOKIE_KEY             = "token"
	PATH_CALLBACK              = "/auth/callback"
	PATH_LOGIN                 = "/auth/login"
)

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
	case PATH_LOGIN:
		s.HandleLogin(w, r)
		return
	case PATH_CALLBACK:
		s.HandleCallback(w, r)
		return
	default:
		// check auth, if its not logged in redirect to /auth/login
		tokenCookie, _ := r.Cookie(JWT_COOKIE_KEY)
		email, _ := s.GetEmailFromJWTToken(tokenCookie)
		if email == "" {
			http.Redirect(w, r, fmt.Sprintf(PATH_LOGIN+"?redirect=%s", r.URL.Path), http.StatusFound)
			return
		}

		fmt.Printf("Authenticated as %q to call: %q\n", email, r.URL.Path)
	}
}

func (s *SSOClient) GetEmailFromJWTToken(c *http.Cookie) (string, error) {
	if c == nil {
		return "", errors.New("no cookie found")
	}
	token, err := jwt.Parse(c.Value, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return HMAC_JWT_SECRET, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if email, ok := claims[JWT_COOKIE_EMAIL_KEY]; ok {
			return email.(string), nil
		}
	}

	return "", errors.New("not valid jwt token")
}

func (s *SSOClient) CreateJWTToken(email string) (string, error) {
	jwtClaims := jwt.MapClaims{
		JWT_COOKIE_EMAIL_KEY: email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	tokenString, err := token.SignedString(HMAC_JWT_SECRET)
	if err != nil {
		log.Errorf("Failed to create JWT token: %q", err.Error())
		return "", err
	}
	return tokenString, nil
}

func (s *SSOClient) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state := Oauth2State{
		Redirect: r.FormValue(FORM_REDIRECT_KEY),
	}
	stateStr, _ := json.Marshal(state)
	stateB64 := base64.StdEncoding.EncodeToString([]byte(string(stateStr)))
	url := s.oAuth2Config.AuthCodeURL(string(stateB64))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *SSOClient) HandleCallback(w http.ResponseWriter, r *http.Request) {
	stateB64 := r.FormValue(FORM_STATE_KEY)
	stateStr, _ := base64.StdEncoding.DecodeString(stateB64)
	state := &Oauth2State{}
	json.Unmarshal([]byte(stateStr), &state)

	code := r.FormValue(FORM_CODE_KEY)
	data, err := s.getUserData(code)
	if err != nil {
		log.Errorf("Error when getting user data: %q", err.Error())
		w.Write([]byte(FAILED_LOGIN_ERROR_MESSAGE))
		return
	}

	jwtTokenCookiePath := "/"
	jwtTokenExpireAt := time.Now().Add(24 * time.Hour)
	jwtTokenCookieValue, err := s.CreateJWTToken(data.Email)
	if err != nil {
		w.Write([]byte(FAILED_LOGIN_ERROR_MESSAGE))
		return
	}
	s.setJWTTokenCookie(w, JWT_COOKIE_KEY, jwtTokenCookieValue, jwtTokenCookiePath, jwtTokenExpireAt)

	http.Redirect(w, r, state.Redirect, http.StatusFound)
}

func (s *SSOClient) getUserData(code string) (*OAuthData, error) {
	token, err := s.oAuth2Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	response, err := http.Get(GOOGLE_API_URL + token.AccessToken)
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

func (s *SSOClient) setJWTTokenCookie(w http.ResponseWriter, cookieName, cookievalue, cookiePath string, expireAt time.Time) {
	cookie := http.Cookie{Name: cookieName, Value: cookievalue, Path: cookiePath, Expires: expireAt}
	http.SetCookie(w, &cookie)
}
