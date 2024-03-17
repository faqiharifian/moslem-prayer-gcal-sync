package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"

	"github.com/faqiharifian/moslem-prayer-gcal-sync/constant"
)

var defaultConfig Config

type Config struct {
	Host   string `envconfig:"HOST" default:"localhost:8085"`
	Oauth2 *oauth2.Config
}

func Load() error {
	err := envconfig.Process("", &defaultConfig)
	if err != nil {
		return err
	}

	return err
}

func (c *Config) LoadOauthConfig(filepath string) error {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	c.Oauth2, err = google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	c.Oauth2.RedirectURL = "http://" + c.Host + constant.CallbackPath
	return nil
}

func Get() Config {
	return defaultConfig
}
