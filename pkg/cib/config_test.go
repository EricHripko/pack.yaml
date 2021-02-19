package cib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	// Act
	cfg := NewConfig()

	// Assert
	require.True(t, cfg.Debug)
	require.Empty(t, cfg.Entrypoint)
	require.Empty(t, cfg.Command)
	require.Equal(t, cfg.User, "nobody")
	require.Empty(t, cfg.Other)
}

func TestReadConfig_Valid(t *testing.T) {
	// Arrange
	data := []byte(`
debug: false
entrypoint: ["entrypoint"]
command: ["command"]
user: somebody
go:
    version: "1.12"
`)

	// Act
	cfg, err := ReadConfig(data)

	// Assert
	require.Nil(t, err)
	require.False(t, cfg.Debug)
	require.Equal(t, []string{"entrypoint"}, cfg.Entrypoint)
	require.Equal(t, []string{"command"}, cfg.Command)
	require.Equal(t, "somebody", cfg.User)
	require.Equal(t, map[string]interface{}{
		"go": map[interface{}]interface{}{
			"version": "1.12",
		},
	}, cfg.Other)
}

func TestReadConfig_InvalidYAML(t *testing.T) {
	// Arrange
	data := []byte("!\"%!%")

	// Act
	_, err := ReadConfig(data)

	// Assert
	require.Error(t, err)
}

func TestReadConfig_InvalidTypes(t *testing.T) {
	// Arrange
	data := []byte("debug: nope")

	// Act
	_, err := ReadConfig(data)

	// Assert
	require.Error(t, err)
}
