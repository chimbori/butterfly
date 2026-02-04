package conf

// App-specific configuration structs & data.
// Must live in a package of its own so other packages within the app can depend on it without
// causing a circular dependency.

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

var AppName = "Butterfly"

var BuildTimestamp string

var Config AppConfig

type AppConfig struct {
	DataDir  string `yaml:"-"` // The directory containing `butterfly.yml` is where all data will be stored.
	Database struct {
		Url string `yaml:"url"`
	} `yaml:"database"`
	Web struct {
		Port int `yaml:"port"`
	} `yaml:"web"`
	Dashboard struct {
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
		Pagination struct {
			Limit int `yaml:"limit"`
		} `yaml:"pagination"`
	} `yaml:"dashboard"`
	Logs struct {
		Retention time.Duration `yaml:"retention"`
	} `yaml:"logs"`
	LinkPreviews struct {
		Screenshot struct {
			Timeout time.Duration `yaml:"timeout"`
		} `yaml:"screenshot"`
		Cache struct {
			Enabled      *bool         `yaml:"enabled"`
			TTL          time.Duration `yaml:"ttl"`
			MaxSizeBytes int64         `yaml:"max_size_bytes"`
		} `yaml:"cache"`
	} `yaml:"link-previews"`
	QrCodes struct {
		Cache struct {
			Enabled      *bool         `yaml:"enabled"`
			TTL          time.Duration `yaml:"ttl"`
			MaxSizeBytes int64         `yaml:"max_size_bytes"`
		} `yaml:"cache"`
	} `yaml:"qr-codes"`
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
	if c.Web.Port == 0 {
		c.Web.Port = 9999
	}
	if c.Dashboard.Pagination.Limit == 0 {
		c.Dashboard.Pagination.Limit = 30
	}

	// Cache for Link Previews is enabled by default; only disable it when testing or debugging.
	if c.LinkPreviews.Cache.Enabled == nil {
		enabled := true
		c.LinkPreviews.Cache.Enabled = &enabled
	}
	if c.LinkPreviews.Cache.MaxSizeBytes == 0 {
		c.LinkPreviews.Cache.MaxSizeBytes = 1 * 1024 * 1024 * 1024 // 1GB
	}
	if c.LinkPreviews.Screenshot.Timeout == 0 {
		c.LinkPreviews.Screenshot.Timeout = 20 * time.Second
	}

	// Cache for QR Codes is enabled by default; only disable it when testing or debugging.
	if c.QrCodes.Cache.Enabled == nil {
		enabled := true
		c.QrCodes.Cache.Enabled = &enabled
	}
	if c.QrCodes.Cache.MaxSizeBytes == 0 {
		c.QrCodes.Cache.MaxSizeBytes = 1 * 1024 * 1024 * 1024 // 1GB
	}

	if c.Logs.Retention == 0 {
		c.Logs.Retention = 30 * 24 * time.Hour
	}

	// Print warnings for unsafe settings, just as FYI.
	yaml, _ := yaml.Marshal(*c)
	fmt.Println(string(yaml))
	if c.Debug {
		slog.Warn("Debug mode is enabled")
	}

	if !*c.LinkPreviews.Cache.Enabled {
		slog.Warn("Screenshot cache disabled for Link Previews; performance will be affected")
	}
	if !*c.QrCodes.Cache.Enabled {
		slog.Warn("Cache disabled for QR Codes; performance will be affected")
	}
}
