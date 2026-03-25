package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// AppName is the directory name used under XDG base directories.
const AppName = "cmdvault"

// Config holds the application configuration.
type Config struct {
	Clipboard      bool   `mapstructure:"clipboard"`
	Editor         string `mapstructure:"editor"`
	DatabasePath   string `mapstructure:"database_path"`
	Color          string `mapstructure:"color"`
	ConfirmExecute bool   `mapstructure:"confirm_execute"`
	DefaultFormat  string `mapstructure:"default_format"`
}

// XDGConfigHome returns $XDG_CONFIG_HOME or ~/.config as fallback.
func XDGConfigHome() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: resolve home directory: %w", err)
	}
	return filepath.Join(home, ".config"), nil
}

// XDGDataHome returns $XDG_DATA_HOME or ~/.local/share as fallback.
func XDGDataHome() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: resolve home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share"), nil
}

// DefaultDatabasePath returns the default database path under XDG_DATA_HOME.
func DefaultDatabasePath() (string, error) {
	dataHome, err := XDGDataHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataHome, AppName, "cmdvault.db"), nil
}

// BindFlags registers persistent flags on a pflag.FlagSet (if not already
// registered) and binds them to Viper keys. Pass nil to skip flag binding.
func BindFlags(flags *pflag.FlagSet) error {
	if flags == nil {
		return nil
	}

	registerBool := func(name string, value bool, usage string) {
		if flags.Lookup(name) == nil {
			flags.Bool(name, value, usage)
		}
	}
	registerString := func(name, value, usage string) {
		if flags.Lookup(name) == nil {
			flags.String(name, value, usage)
		}
	}

	registerBool("clipboard", true, "copy snippets to clipboard")
	registerString("editor", "", "editor command for editing snippets")
	registerString("database-path", "", "path to the SQLite database file")
	registerString("color", "auto", "color output mode (auto, always, never)")
	registerBool("confirm-execute", true, "prompt before executing snippets")
	registerString("default-format", "yaml", "default export format (json, yaml)")

	bindings := []struct {
		key  string
		flag string
	}{
		{"clipboard", "clipboard"},
		{"editor", "editor"},
		{"database_path", "database-path"},
		{"color", "color"},
		{"confirm_execute", "confirm-execute"},
		{"default_format", "default-format"},
	}

	for _, b := range bindings {
		if err := viper.BindPFlag(b.key, flags.Lookup(b.flag)); err != nil {
			return fmt.Errorf("config: bind flag %s: %w", b.flag, err)
		}
	}

	return nil
}

// InitConfig loads configuration from file, environment, and defaults.
// Call BindFlags before this if you want CLI flags to participate in the
// precedence chain (flags > env > config file > defaults).
func InitConfig(flags *pflag.FlagSet, configPath string) (*Config, error) {
	var cfg Config

	dbPath, err := DefaultDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("config: resolve default database path: %w", err)
	}

	viper.SetDefault("clipboard", true)
	viper.SetDefault("editor", "$EDITOR")
	viper.SetDefault("database_path", dbPath)
	viper.SetDefault("color", "auto")
	viper.SetDefault("confirm_execute", true)
	viper.SetDefault("default_format", "yaml")

	viper.SetEnvPrefix("CMDVAULT")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		configHome, err := XDGConfigHome()
		if err != nil {
			return nil, fmt.Errorf("config: resolve config directory: %w", err)
		}
		viper.AddConfigPath(filepath.Join(configHome, AppName))

		fmt.Println(configHome)
	}

	viper.AutomaticEnv()

	if err := BindFlags(flags); err != nil {
		return nil, err
	}

	if err := viper.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil, fmt.Errorf("config: read config file: %w", err)
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}

	return &cfg, nil
}
