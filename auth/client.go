package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"

	"github.com/faqiharifian/moslem-prayer-gcal-sync/config"
)

const tokenFilePath = ".token.json"

type Client struct {
	cfg *oauth2.Config
}

func NewClient(cfg config.Config) Client {
	return Client{cfg: cfg.Oauth2}
}

func (c *Client) Auth(state string, tokCh chan *oauth2.Token) {
	tok, err := c.tokenFromFile()
	if err == nil {
		tokCh <- tok
		return
	}

	authURL := c.cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)

	fmt.Println("Please give google calendar permission to this app to continue. Opening browser...")
	fmt.Println()
	fmt.Println(authURL)
	fmt.Println()

	time.Sleep(2 * time.Second)

	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", authURL).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", authURL).Start()
	case "darwin":
		exec.Command("open", authURL).Start()
	}
}

func (c *Client) UpdateToken(tok *oauth2.Token) {
	f, err := os.OpenFile(tokenFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(tok)
}

func (c *Client) tokenFromFile() (*oauth2.Token, error) {
	f, err := os.Open(tokenFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	if err != nil {
		return nil, err
	}
	if time.Now().After(tok.Expiry) {
		return nil, errors.New("token expired")
	}
	return tok, nil
}
