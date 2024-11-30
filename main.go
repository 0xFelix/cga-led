package main

import (
	"context"
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
	SaltWebUI string `json:"saltwebui"`
}

type setDeviceResponse struct {
	Error   string        `json:"error"`
	Message string        `json:"message"`
	Token   string        `json:"token"`
	Data    setDeviceData `json:"data"`
}

type setDeviceData struct {
	Led       string `json:"led"`
	HTTPState string `json:"http_state"`
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

type cgaLed struct {
	client *http.Client

	address  string
	username string
	password string
	led      bool
}

func (c *cgaLed) init() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	c.client = &http.Client{Jar: jar}
	c.parseFlags()
}

func (c *cgaLed) parseFlags() {
	flag.StringVar(&c.address, "a", "192.168.100.1", "Address of API")
	flag.StringVar(&c.password, "p", "password", "Password for API")
	flag.StringVar(&c.username, "u", "admin", "Username for API")
	flag.BoolVar(&c.led, "l", false, "Turn led on (true) or off (false)")
	flag.Parse()
}

func (c *cgaLed) login() error {
	res, err := c.sendSessionLogin("seeksalthash")
	if err != nil {
		return err
	}

	if _, err := c.sendSessionLogin(deriveChallenge(res, c.password)); err != nil {
		return err
	}

	return c.sendSessionMenu()
}

func (c *cgaLed) logout() error {
	res, err := c.sendHostHostTbl()
	if err != nil {
		return err
	}

	return c.sendSessionLogout(res.Token)
}

func (c *cgaLed) setLed() error {
	res, err := c.sendSetDevice()
	if err != nil {
		return err
	}

	current, err := strconv.ParseBool(res.Data.Led)
	if err != nil {
		return err
	}

	if current != c.led {
		return c.sendSetDeviceSdevice(res.Data.HTTPState, res.Token)
	}

	return nil
}

func (c *cgaLed) sendHostHostTbl() (*genericResponse, error) {
	return sendRequest[genericResponse](
		c.client,
		http.MethodGet,
		fmt.Sprintf("http://%s/api/v1/host/hostTbl", c.address),
		nil,
		"",
	)
}

func (c *cgaLed) sendSessionLogin(password string) (*loginResponse, error) {
	values := url.Values{}
	values.Set("username", c.username)
	values.Set("password", password)
	// By setting this value an already active session is logged out before the new login attempt
	values.Set("logout", "true")

	return sendRequest[loginResponse](
		c.client,
		http.MethodPost,
		fmt.Sprintf("http://%s/api/v1/session/login", c.address),
		strings.NewReader(values.Encode()),
		"",
	)
}

func (c *cgaLed) sendSessionLogout(token string) error {
	_, err := sendRequest[genericResponse](
		c.client,
		http.MethodPost,
		fmt.Sprintf("http://%s/api/v1/session/logout", c.address),
		nil,
		token,
	)
	return err
}

func (c *cgaLed) sendSessionMenu() error {
	_, err := sendRequest[genericResponse](
		c.client,
		http.MethodGet,
		fmt.Sprintf("http://%s/api/v1/session/menu", c.address),
		nil,
		"",
	)
	return err
}

func (c *cgaLed) sendSetDevice() (*setDeviceResponse, error) {
	return sendRequest[setDeviceResponse](
		c.client,
		http.MethodGet,
		fmt.Sprintf("http://%s/api/v1/set_device", c.address),
		nil,
		"",
	)
}

func (c *cgaLed) sendSetDeviceSdevice(httpState, token string) error {
	values := url.Values{}
	values.Set("led", strconv.FormatBool(c.led))
	values.Set("http_state", httpState)

	_, err := sendRequest[genericResponse](
		c.client,
		http.MethodPost,
		fmt.Sprintf("http://%s/api/v1/set_device/Sdevice", c.address),
		strings.NewReader(values.Encode()),
		token,
	)
	return err
}

func sendRequest[T apiResponse](client *http.Client, method, u string, body io.Reader, token string) (*T, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, u, body)
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

func deriveChallenge(res *loginResponse, password string) string {
	const iterations = 1000
	const keyLen = 16
	a := pbkdf2.Key([]byte(password), []byte(res.Salt), iterations, keyLen, sha256.New)
	b := pbkdf2.Key([]byte(hex.EncodeToString(a)), []byte(res.SaltWebUI), iterations, keyLen, sha256.New)
	return hex.EncodeToString(b)
}

func main() {
	c := cgaLed{}
	c.init()

	if err := c.login(); err != nil {
		log.Fatal(err)
	}

	if err := c.setLed(); err != nil {
		log.Fatal(err)
	}

	if err := c.logout(); err != nil {
		log.Fatal(err)
	}
}
