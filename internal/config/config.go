package config

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/dlazz/windows-management-rest/internal/module"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

const (
	WMR_MODULES         = "WMR_MODULES"
	WMR_TOKEN           = "WMR_TOKEN"
	WMR_WEBSERVER_PORT  = "WMR_WEBSERVER_PORT"
	WMR_WEBSERVER_DEBUG = "WMR_WEBSERVER_DEBUG"
)

var Manager *Configuration

type Configuration struct {
	Webserver *Webserver `json:"webserver"`
	Token     string     `json:"auth_token"`
	Modules   []string   `json:"modules"`
}

func (c *Configuration) validate() {
	c.Webserver.validate()
	if c.Token == "" && os.Getenv(WMR_TOKEN) == "" {
		log.Error().Str("configuration", "token").Msg("a valid authentication token must be set")
		os.Exit(1)
	}
	if c.Token == "" {
		c.Token = os.Getenv(WMR_TOKEN)
	}
	if err := c.HashToken(); err != nil {
		log.Error().Str("configuration", "token").Msg("invalid token configuration")
		os.Exit(1)
	}
	if c.Modules == nil && os.Getenv(WMR_MODULES) == "" {
		log.Error().Str("configuration", "modules").Msg("at least a valid module must be set")
		os.Exit(1)
	}
	if c.Modules == nil {
		c.Modules = strings.Split(os.Getenv(WMR_MODULES), ",")
	}
	c.Modules = loadModule(c.Modules)
}

type Webserver struct {
	Debug bool   `json:"debug"`
	Port  string `json:"port"`
}

func (w *Webserver) validate() {
	if w.Port == "" && os.Getenv(WMR_WEBSERVER_PORT) == "" {
		w.Port = "9898"
	}
	if w.Port == "" {
		w.Port = os.Getenv(WMR_WEBSERVER_PORT)
	}
	if _, err := strconv.Atoi(w.Port); err != nil {
		log.Error().Str("configuration", "webserver").Msg(w.Port + " is not a valid webserver port")
		os.Exit(1)
	}
	dbg := os.Getenv(WMR_WEBSERVER_DEBUG)
	if dbg != "" {
		debug, err := strconv.ParseBool(dbg)
		if err != nil {
			log.Error().Err(err).Str("configuration", "webserver").Msg(dbg + " is not a valid webserver port")
		} else {
			w.Debug = debug
		}
	}
}

func loadJson(r io.Reader) *Configuration {
	c := &Configuration{}
	if err := json.NewDecoder(r).Decode(c); err != nil {
		log.Error().Err(err).Msg("unable to read decode json")
		os.Exit(1)
	}
	return c
}

func loadModule(mod []string) []string {
	toLoad := []string{}
	for _, m := range mod {
		if _, ok := module.Store[m]; ok {
			toLoad = append(toLoad, m)
		}
	}
	return toLoad
}

func (c *Configuration) HashToken() error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(c.Token), 14)
	if err != nil {
		return err
	}
	c.Token = string(bytes)
	return nil
}

func InitConfig(r io.Reader) {
	Manager = loadJson(r)
	Manager.validate()
}
