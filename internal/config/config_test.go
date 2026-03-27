package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(func() { viper.Reset() })
}

func TestXDGConfigHome(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name     string
		envValue string
		want     string
	}{
		{
			name:     "uses XDG_CONFIG_HOME when set",
			envValue: "/custom/config",
			want:     "/custom/config",
		},
		{
			name:     "falls back to ~/.config when unset",
			envValue: "",
			want:     filepath.Join(home, ".config"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prev, hadPrev := os.LookupEnv("XDG_CONFIG_HOME")
			if tt.envValue != "" {
				os.Setenv("XDG_CONFIG_HOME", tt.envValue)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
			}
			t.Cleanup(func() {
				if hadPrev {
					os.Setenv("XDG_CONFIG_HOME", prev)
				} else {
					os.Unsetenv("XDG_CONFIG_HOME")
				}
			})

			got, err := XDGConfigHome()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestXDGDataHome(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name     string
		envValue string
		want     string
	}{
		{
			name:     "uses XDG_DATA_HOME when set",
			envValue: "/custom/data",
			want:     "/custom/data",
		},
		{
			name:     "falls back to ~/.local/share when unset",
			envValue: "",
			want:     filepath.Join(home, ".local", "share"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prev, hadPrev := os.LookupEnv("XDG_DATA_HOME")
			if tt.envValue != "" {
				os.Setenv("XDG_DATA_HOME", tt.envValue)
			} else {
				os.Unsetenv("XDG_DATA_HOME")
			}
			t.Cleanup(func() {
				if hadPrev {
					os.Setenv("XDG_DATA_HOME", prev)
				} else {
					os.Unsetenv("XDG_DATA_HOME")
				}
			})

			got, err := XDGDataHome()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDefaultDatabasePath(t *testing.T) {
	t.Run("uses XDG_DATA_HOME", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/tmp/testdata")

		got, err := DefaultDatabasePath()
		require.NoError(t, err)
		assert.Equal(t, "/tmp/testdata/cmdvault/cmdvault.db", got)
	})

	t.Run("falls back to home directory", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "")
		os.Unsetenv("XDG_DATA_HOME")

		home, err := os.UserHomeDir()
		require.NoError(t, err)

		got, err := DefaultDatabasePath()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".local", "share", "cmdvault", "cmdvault.db"), got)
	})
}

func TestInitConfig_Defaults(t *testing.T) {
	resetViper(t)

	// Point XDG to empty temp dir so no real config file is discovered
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := InitConfig(nil, "")
	require.NoError(t, err)

	assert.True(t, cfg.Clipboard)
	assert.Equal(t, "nano", cfg.Editor)
	assert.Equal(t, "auto", cfg.Color)
	assert.True(t, cfg.ConfirmExecute)
	assert.Equal(t, "yaml", cfg.DefaultFormat)
	assert.Contains(t, cfg.DatabasePath, "cmdvault.db")
}

func TestInitConfig_ExplicitConfigFile(t *testing.T) {
	resetViper(t)

	configFile := filepath.Join(t.TempDir(), "custom.yaml")
	err := os.WriteFile(configFile, []byte("color: never\ndefault_format: json\n"), 0o644)
	require.NoError(t, err)

	cfg, err := InitConfig(nil, configFile)
	require.NoError(t, err)

	assert.Equal(t, "never", cfg.Color)
	assert.Equal(t, "json", cfg.DefaultFormat)
	// Defaults still apply for unset values
	assert.True(t, cfg.Clipboard)
}

func TestInitConfig_ConfigFileXDGDiscovery(t *testing.T) {
	resetViper(t)

	configDir := filepath.Join(t.TempDir(), AppName)
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("color: never\ndefault_format: json\n"), 0o644)
	require.NoError(t, err)

	t.Setenv("XDG_CONFIG_HOME", filepath.Dir(configDir))

	cfg, err := InitConfig(nil, "")
	require.NoError(t, err)

	assert.Equal(t, "never", cfg.Color)
	assert.Equal(t, "json", cfg.DefaultFormat)
	assert.True(t, cfg.Clipboard)
}

func TestInitConfig_EnvOverridesFile(t *testing.T) {
	resetViper(t)

	configDir := filepath.Join(t.TempDir(), AppName)
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("color: never\n"), 0o644)
	require.NoError(t, err)

	t.Setenv("XDG_CONFIG_HOME", filepath.Dir(configDir))
	t.Setenv("CMDVAULT_COLOR", "always")

	cfg, err := InitConfig(nil, "")
	require.NoError(t, err)

	assert.Equal(t, "always", cfg.Color)
}

func TestInitConfig_FlagsOverrideAll(t *testing.T) {
	resetViper(t)

	configDir := filepath.Join(t.TempDir(), AppName)
	err := os.MkdirAll(configDir, 0o755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("color: never\n"), 0o644)
	require.NoError(t, err)

	t.Setenv("XDG_CONFIG_HOME", filepath.Dir(configDir))
	t.Setenv("CMDVAULT_COLOR", "always")

	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	cfg, err := InitConfig(flags, "")
	require.NoError(t, err)

	// Flag not explicitly set, so env wins
	assert.Equal(t, "always", cfg.Color)

	// Now reset and set the flag explicitly
	viper.Reset()
	flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	// Pre-register flags so we can parse before InitConfig
	flags.Bool("clipboard", true, "")
	flags.String("editor", "", "")
	flags.String("database-path", "", "")
	flags.String("color", "auto", "")
	flags.Bool("confirm-execute", true, "")
	flags.String("default-format", "yaml", "")
	err = flags.Parse([]string{"--color", "auto"})
	require.NoError(t, err)

	cfg, err = InitConfig(flags, "")
	require.NoError(t, err)

	assert.Equal(t, "auto", cfg.Color)
}

func TestBindFlags_Nil(t *testing.T) {
	// Should not panic with nil flags
	assert.NotPanics(t, func() {
		err := BindFlags(nil)
		assert.NoError(t, err)
	})
}

func TestBindFlags_RegistersAllFlags(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	err := BindFlags(flags)
	require.NoError(t, err)

	expectedFlags := []string{
		"clipboard",
		"editor",
		"database-path",
		"color",
		"confirm-execute",
		"default-format",
	}

	for _, name := range expectedFlags {
		assert.NotNil(t, flags.Lookup(name), "flag %q should be registered", name)
	}
}
