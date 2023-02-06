package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

type genericResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

type loginResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message,omitempty"`
	Salt      string `json:"salt"`
	SaltWebUi string `json:"saltwebui"`
}

type setDeviceResponse struct {
	Error   string        `json:"error"`
	Message string        `json:"message"`
	Token   string        `json:"token"`
	Data    setDeviceData `json:"data"`
}

type setDeviceData struct {
	Led       string `json:"led"`
	HttpState string `json:"http_state"`
}

type apiResponse interface {
	GetError() string
	GetMessage() string
}

func (r genericResponse) GetError() string {
	return r.Error
}

func (r genericResponse) GetMessage() string {
	return r.Message
}

func (r loginResponse) GetError() string {
	return r.Error
}

func (r loginResponse) GetMessage() string {
	return r.Message
}

func (r setDeviceResponse) GetError() string {
	return r.Error
}

func (r setDeviceResponse) GetMessage() string {
	return r.Message
}

var (
	address  = "192.168.100.1"
	username = "admin"
	password = "password"
	led      = false
)

func init() {
	flag.StringVar(&address, "a", "192.168.100.1", "Address of API")
	flag.StringVar(&password, "p", "password", "Password for API")
	flag.StringVar(&username, "u", "admin", "Username for API")
	flag.BoolVar(&led, "l", false, "Turn led on (true) or off (false)")
	flag.Parse()
}

func main() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{Jar: jar}

	err = login(client)
	if err != nil {
		log.Fatal(err)
	}

	err = setLed(client)
	if err != nil {
		log.Fatal(err)
	}

	err = logout(client)
	if err != nil {
		log.Fatal(err)
	}
}

func login(client *http.Client) error {
	res, err := sendSessionLogin(client, "seeksalthash")
	if err != nil {
		return err
	}

	_, err = sendSessionLogin(client, deriveChallenge(res, password))
	if err != nil {
		return err
	}

	err = sendSessionMenu(client)
	if err != nil {
		return err
	}

	return nil
}

func logout(client *http.Client) error {
	res, err := sendHostHostTbl(client)
	if err != nil {
		return err
	}

	return sendSessionLogout(client, res.Token)
}

func setLed(client *http.Client) error {
	res, err := sendSetDevice(client)
	if err != nil {
		return err
	}

	current, err := strconv.ParseBool(res.Data.Led)
	if err != nil {
		return err
	}

	if current != led {
		return sendSetDeviceSdevice(client, strconv.FormatBool(led), res.Data.HttpState, res.Token)
	}

	return nil
}

func sendRequest[T apiResponse](client *http.Client, method string, url string, body io.Reader, token string) (*T, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	if token != "" {
		req.Header.Set("X-Csrf-Token", token)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = res.Body.Close()
	if err != nil {
		return nil, err
	}

	var apiRes T
	err = json.Unmarshal(data, &apiRes)
	if err != nil {
		return nil, err
	}

	if apiRes.GetError() != "ok" {
		return nil, fmt.Errorf("api response was not ok: error '%s', message '%s'", apiRes.GetError(), apiRes.GetMessage())
	}

	return &apiRes, nil
}

func deriveChallenge(loginRes *loginResponse, password string) string {
	a := pbkdf2.Key([]byte(password), []byte(loginRes.Salt), 1000, 16, sha256.New)
	b := pbkdf2.Key([]byte(hex.EncodeToString(a)), []byte(loginRes.SaltWebUi), 1000, 16, sha256.New)
	return hex.EncodeToString(b)
}

func sendHostHostTbl(client *http.Client) (*genericResponse, error) {
	return sendRequest[genericResponse](client, http.MethodGet, fmt.Sprintf("http://%s/api/v1/host/hostTbl", address), nil, "")
}

func sendSessionLogin(client *http.Client, password string) (*loginResponse, error) {
	values := url.Values{}
	values.Set("username", username)
	values.Set("password", password)
	// By setting this value an already active session is logged out before the new login attempt
	values.Set("logout", "true")

	return sendRequest[loginResponse](client, http.MethodPost, fmt.Sprintf("http://%s/api/v1/session/login", address), strings.NewReader(values.Encode()), "")
}

func sendSessionLogout(client *http.Client, token string) error {
	_, err := sendRequest[genericResponse](client, http.MethodPost, fmt.Sprintf("http://%s/api/v1/session/logout", address), nil, token)
	if err != nil {
		return err
	}

	return nil
}

func sendSessionMenu(client *http.Client) error {
	_, err := sendRequest[genericResponse](client, http.MethodGet, fmt.Sprintf("http://%s/api/v1/session/menu", address), nil, "")
	if err != nil {
		return err
	}

	return nil
}

func sendSetDevice(client *http.Client) (*setDeviceResponse, error) {
	return sendRequest[setDeviceResponse](client, http.MethodGet, fmt.Sprintf("http://%s/api/v1/set_device", address), nil, "")
}

func sendSetDeviceSdevice(client *http.Client, led string, httpState string, token string) error {
	values := url.Values{}
	values.Set("led", led)
	values.Set("http_state", httpState)

	_, err := sendRequest[genericResponse](client, http.MethodPost, fmt.Sprintf("http://%s/api/v1/set_device/Sdevice", address), strings.NewReader(values.Encode()), token)
	if err != nil {
		return err
	}

	return nil
}
