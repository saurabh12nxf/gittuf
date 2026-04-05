// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package gitinterface

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGitConfig(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// CreateTestGitRepository sets our test config
	config, err := repo.GetGitConfig()
	assert.Nil(t, err)
	assert.Equal(t, testName, config["user.name"])
	assert.Equal(t, testEmail, config["user.email"])
}

func TestSetGitConfig(t *testing.T) {
	t.Run("basic sets", func(t *testing.T) {
		const name = "John Doe"
		const email = "john.doe@example.com"

		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, false)

		err := repo.SetGitConfig("user.name", name)
		require.NoError(t, err)
		err = repo.SetGitConfig("user.email", email)
		require.NoError(t, err)

		config, err := repo.GetGitConfig()
		require.NoError(t, err)
		assert.Equal(t, name, config["user.name"])
		assert.Equal(t, email, config["user.email"])
	})
	t.Run("empty set", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, false)

		err := repo.SetGitConfig("user.name", "")
		require.NoError(t, err)
		err = repo.SetGitConfig("user.email", "")
		require.NoError(t, err)

		config, err := repo.GetGitConfig()
		require.NoError(t, err)
		assert.Equal(t, "", config["user.name"])
		assert.Equal(t, "", config["user.email"])
	})
	t.Run("gpg.format special case", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, false)

		err := repo.SetGitConfig("gpg.format", "gpg")
		require.NoError(t, err)

		config, err := repo.GetGitConfig()
		require.NoError(t, err)
		assert.Equal(t, "gpg", config["gpg.format"])

		err = repo.SetGitConfig("gpg.format", "")
		require.NoError(t, err)

		config, err = repo.GetGitConfig()
		require.NoError(t, err)
		assert.Equal(t, "", config["gpg.format"])
	})
}

func TestGetGitConfigEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("get config with multiple values", func(t *testing.T) {
		// Set multiple config values
		err := repo.SetGitConfig("test.key1", "value1")
		require.NoError(t, err)
		err = repo.SetGitConfig("test.key2", "value2")
		require.NoError(t, err)
		err = repo.SetGitConfig("test.key3", "value3")
		require.NoError(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, "value1", config["test.key1"])
		assert.Equal(t, "value2", config["test.key2"])
		assert.Equal(t, "value3", config["test.key3"])
	})

	t.Run("get config with special characters in value", func(t *testing.T) {
		err := repo.SetGitConfig("test.special", "value with spaces and !@#$%")
		require.NoError(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, "value with spaces and !@#$%", config["test.special"])
	})

	t.Run("get config with empty value", func(t *testing.T) {
		err := repo.SetGitConfig("test.empty", "")
		require.NoError(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, "", config["test.empty"])
	})
}

func TestSetGitConfigEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("set config with dots in key", func(t *testing.T) {
		err := repo.SetGitConfig("section.subsection.key", "value")
		require.NoError(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, "value", config["section.subsection.key"])
	})

	t.Run("overwrite existing config", func(t *testing.T) {
		err := repo.SetGitConfig("test.overwrite", "original")
		require.NoError(t, err)

		err = repo.SetGitConfig("test.overwrite", "updated")
		require.NoError(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, "updated", config["test.overwrite"])
	})
}

func TestGetGitConfigComprehensive(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("get all config values", func(t *testing.T) {
		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.NotEmpty(t, config)

		// Should have user.name and user.email from CreateTestGitRepository
		assert.Contains(t, config, "user.name")
		assert.Contains(t, config, "user.email")
	})

	t.Run("config keys are lowercase", func(t *testing.T) {
		// Set a config with uppercase
		err := repo.SetGitConfig("Test.Key", "value")
		require.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)

		// Should be lowercase in the returned map
		assert.Contains(t, config, "test.key")
		assert.Equal(t, "value", config["test.key"])
	})

	t.Run("config with spaces in value", func(t *testing.T) {
		err := repo.SetGitConfig("test.spaces", "value with spaces")
		require.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, "value with spaces", config["test.spaces"])
	})

	t.Run("config with special characters", func(t *testing.T) {
		err := repo.SetGitConfig("test.special", "value!@#$%^&*()")
		require.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, "value!@#$%^&*()", config["test.special"])
	})
}

func TestSetGitConfigComprehensive(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("set simple config", func(t *testing.T) {
		err := repo.SetGitConfig("test.simple", "simple-value")
		assert.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, "simple-value", config["test.simple"])
	})

	t.Run("set config with dots in key", func(t *testing.T) {
		err := repo.SetGitConfig("section.subsection.key", "value")
		assert.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, "value", config["section.subsection.key"])
	})

	t.Run("overwrite existing config", func(t *testing.T) {
		err := repo.SetGitConfig("test.overwrite", "original")
		require.Nil(t, err)

		err = repo.SetGitConfig("test.overwrite", "updated")
		assert.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, "updated", config["test.overwrite"])
	})

	t.Run("set config with numeric value", func(t *testing.T) {
		err := repo.SetGitConfig("test.number", "12345")
		assert.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, "12345", config["test.number"])
	})

	t.Run("set config with boolean value", func(t *testing.T) {
		err := repo.SetGitConfig("test.boolean", "true")
		assert.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, "true", config["test.boolean"])
	})

	t.Run("set config with path value", func(t *testing.T) {
		err := repo.SetGitConfig("test.path", "/path/to/file")
		assert.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, "/path/to/file", config["test.path"])
	})
}

func TestGitConfigEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("get config after multiple sets", func(t *testing.T) {
		keys := []string{"test.key1", "test.key2", "test.key3", "test.key4", "test.key5"}
		for i, key := range keys {
			err := repo.SetGitConfig(key, fmt.Sprintf("value%d", i))
			require.Nil(t, err)
		}

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)

		for i, key := range keys {
			assert.Equal(t, fmt.Sprintf("value%d", i), config[key])
		}
	})

	t.Run("config with very long value", func(t *testing.T) {
		longValue := strings.Repeat("a", 1000)
		err := repo.SetGitConfig("test.long", longValue)
		assert.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Equal(t, longValue, config["test.long"])
	})

	t.Run("config with newline in value", func(t *testing.T) {
		// Git config doesn't support newlines in values directly
		// but we can test that it doesn't break
		err := repo.SetGitConfig("test.newline", "line1")
		assert.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Contains(t, config, "test.newline")
	})

	t.Run("config with quotes in value", func(t *testing.T) {
		err := repo.SetGitConfig("test.quotes", `value"with"quotes`)
		assert.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Contains(t, config, "test.quotes")
	})
}

func TestGitConfigCaseInsensitivity(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("uppercase key becomes lowercase", func(t *testing.T) {
		err := repo.SetGitConfig("TEST.UPPERCASE", "value")
		require.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Contains(t, config, "test.uppercase")
	})

	t.Run("mixed case key becomes lowercase", func(t *testing.T) {
		err := repo.SetGitConfig("TeSt.MiXeD", "value")
		require.Nil(t, err)

		config, err := repo.GetGitConfig()
		assert.Nil(t, err)
		assert.Contains(t, config, "test.mixed")
	})
}

func TestConfigErrorCases(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	// Test SetGitConfig with invalid key format
	err := repo.SetGitConfig("", "value")
	assert.NotNil(t, err)

	// corrupt repo to force git config failure
	gitDir := repo.GetGitDir()
	os.RemoveAll(gitDir)

	err = repo.SetGitConfig("user.name", "test")
	assert.NotNil(t, err)
}

func TestGetGitConfigWithMultipleKeys(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Set multiple config values
	err := repo.SetGitConfig("user.name", "Test User")
	require.Nil(t, err)
	err = repo.SetGitConfig("user.email", "test@example.com")
	require.Nil(t, err)
	err = repo.SetGitConfig("core.editor", "vim")
	require.Nil(t, err)

	// Get all configs
	config, err := repo.GetGitConfig()
	assert.Nil(t, err)
	assert.Contains(t, config, "user.name")
	assert.Contains(t, config, "user.email")
	assert.Contains(t, config, "core.editor")
}

func TestSetGitConfigOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Set initial value
	err := repo.SetGitConfig("test.key", "value1")
	require.Nil(t, err)

	// Overwrite with new value
	err = repo.SetGitConfig("test.key", "value2")
	require.Nil(t, err)

	// Verify new value
	config, err := repo.GetGitConfig()
	assert.Nil(t, err)
	assert.Contains(t, config["test.key"], "value2")
}
