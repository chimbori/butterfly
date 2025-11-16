package conf

// App-specific configuration structs & data.
// Must live in a package of its own so other packages within the app can depend on it without
// causing a circular dependency.

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
	"go.chimbori.app/butterfly/core"
	"gopkg.in/yaml.v3"
)

var AppName = "Butterfly"

var BuildTimestamp string

var Config *AppConfig

type AppConfig struct {
	DataDir     string // The directory containing `butterfly.yml` is where all data will be stored.
	LinkPreview struct {
		Domains []string `yaml:"domains"`
	} `yaml:"link-preview"`
	Web struct {
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		PublicUrl string `yaml:"public-url"`
	} `yaml:"web"`
	Debug bool `yaml:"debug"`
}

func ReadConfig(configYmlFile string) (*AppConfig, error) {
	if BuildTimestamp == "" {
		BuildTimestamp = time.Now().Local().Format("2006-01-02 15:04:05")
	}

	configYmlPath, err := filepath.Abs(configYmlFile)
	if err != nil {
		return nil, err
	}

	buf, err := os.ReadFile(configYmlPath)
	if err != nil {
		return nil, err
	}

	c := &AppConfig{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		slog.Error("Failed to parse config", tint.Err(err))
		return nil, err
	}

	c.DataDir = filepath.Dir(configYmlPath)

	if len(c.LinkPreview.Domains) == 0 {
		return nil, fmt.Errorf("Must provide a list of allowed domains in link-preview.domains")
	}

	if c.Web.Host == "" {
		// Don’t replace this by string(…); the net.IP --> string conversion will fail.
		c.Web.Host = fmt.Sprintf("%s", core.GetOutboundIP())
	}
	if c.Web.PublicUrl == "" {
		c.Web.PublicUrl = "/"
	}

	json, _ := json.MarshalIndent(*c, "", "\t")
	slog.Info("Config read")
	fmt.Println(string(json))
	if c.Debug {
		slog.Warn("Debug mode is enabled")
	}

	return c, err
}
