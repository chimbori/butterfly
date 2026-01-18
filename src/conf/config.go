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

	"chimbori.dev/butterfly/core"
	"gopkg.in/yaml.v3"
)

var AppName = "Butterfly"

var BuildTimestamp string

var Config AppConfig

type AppConfig struct {
	DataDir  string // The directory containing `butterfly.yml` is where all data will be stored.
	Database struct {
		Url string `yaml:"url"`
	} `yaml:"database"`
	Web struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"web"`
	Dashboard struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"dashboard"`
	LinkPreview struct {
		Screenshot struct {
			Timeout time.Duration `yaml:"timeout"`
		} `yaml:"screenshot"`
		Cache struct {
			Enabled *bool `yaml:"enabled"`
		} `yaml:"cache"`
	} `yaml:"link-preview"`
	QrCode struct {
		Cache struct {
			Enabled *bool `yaml:"enabled"`
		} `yaml:"cache"`
	} `yaml:"qr-code"`
	Debug bool `yaml:"debug"`
}

var configYmlPath string

func ReadConfig(configYmlFile string) (AppConfig, error) {
	if BuildTimestamp == "" {
		BuildTimestamp = time.Now().Local().Format("2006-01-02 15:04:05")
	}

	c := &AppConfig{}
	var err error
	configYmlPath, err = filepath.Abs(configYmlFile)
	if err != nil {
		setDefaultsAndPrint(c)
		return *c, fmt.Errorf("Failed to get path to config file: %w", err)
	}

	buf, err := os.ReadFile(configYmlPath)
	if err != nil {
		setDefaultsAndPrint(c)
		return *c, fmt.Errorf("Failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(buf, c)
	if err != nil {
		setDefaultsAndPrint(c)
		return *c, fmt.Errorf("Failed to parse config: %w", err)
	}

	setDefaultsAndPrint(c)
	return *c, err
}

func setDefaultsAndPrint(c *AppConfig) {
	c.DataDir = filepath.Dir(configYmlPath)
	if c.Web.Host == "" {
		// Don’t replace this by string(…); the net.IP --> string conversion will fail.
		c.Web.Host = fmt.Sprintf("%s", core.GetOutboundIP())
	}
	if c.Web.Port == 0 {
		c.Web.Port = 9999
	}

	// Cache for Link Previews is enabled by default; only disable it when testing or debugging.
	if c.LinkPreview.Cache.Enabled == nil {
		enabled := true
		c.LinkPreview.Cache.Enabled = &enabled
	}
	if c.LinkPreview.Screenshot.Timeout == 0 {
		c.LinkPreview.Screenshot.Timeout = 20 * time.Second
	}

	// Cache for QR Codes is enabled by default; only disable it when testing or debugging.
	if c.QrCode.Cache.Enabled == nil {
		enabled := true
		c.QrCode.Cache.Enabled = &enabled
	}

	// Print warnings for unsafe settings, just as FYI.
	json, _ := json.MarshalIndent(*c, "", "\t")
	fmt.Println(string(json))
	if c.Debug {
		slog.Warn("Debug mode is enabled")
	}

	if !*c.LinkPreview.Cache.Enabled {
		slog.Warn("Screenshot cache disabled for Link Previews; performance will be affected")
	}
	if !*c.QrCode.Cache.Enabled {
		slog.Warn("Cache disabled for QR Codes; performance will be affected")
	}
}
