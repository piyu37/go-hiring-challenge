package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRequiresEnvVars(t *testing.T) {
	t.Setenv("HTTP_PORT", "")
	t.Setenv("POSTGRES_PORT", "")
	t.Setenv("POSTGRES_USER", "")
	t.Setenv("POSTGRES_PASSWORD", "")
	t.Setenv("POSTGRES_DB", "")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP_PORT")
}

func TestLoadSuccess(t *testing.T) {
	t.Setenv("HTTP_PORT", "8484")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "postgres")
	t.Setenv("POSTGRES_PASSWORD", "password")
	t.Setenv("POSTGRES_DB", "challenge")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "8484", cfg.HTTPPort)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, defaultRateLimitRPS, cfg.RateLimitRPS)
}
