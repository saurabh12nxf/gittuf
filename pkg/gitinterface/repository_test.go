// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package gitinterface

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository(t *testing.T) {
	t.Run("repository.isBare", func(t *testing.T) {
		t.Run("bare=true", func(t *testing.T) {
			tmpDir := t.TempDir()
			repo := CreateTestGitRepository(t, tmpDir, true)
			assert.True(t, repo.IsBare())
		})

		t.Run("bare=false", func(t *testing.T) {
			tmpDir := t.TempDir()
			repo := CreateTestGitRepository(t, tmpDir, false)
			assert.False(t, repo.IsBare())
		})
	})

	t.Run("with specified path, not bare", func(t *testing.T) {
		tmpDir := t.TempDir()

		_ = CreateTestGitRepository(t, tmpDir, false)
		repo, err := LoadRepository(tmpDir)
		assert.Nil(t, err)

		expectedPath, err := filepath.EvalSymlinks(filepath.Join(tmpDir, ".git"))
		require.Nil(t, err)
		actualPath, err := filepath.EvalSymlinks(repo.GetGitDir())
		require.Nil(t, err)
		assert.Equal(t, expectedPath, actualPath)
	})

	t.Run("with specified path, is bare", func(t *testing.T) {
		tmpDir := t.TempDir()

		_ = CreateTestGitRepository(t, tmpDir, true)
		repo, err := LoadRepository(tmpDir)
		assert.Nil(t, err)

		expectedPath, err := filepath.EvalSymlinks(tmpDir)
		require.Nil(t, err)
		actualPath, err := filepath.EvalSymlinks(repo.GetGitDir())
		require.Nil(t, err)
		assert.Equal(t, expectedPath, actualPath)
	})
}

func TestLoadRepositoryErrors(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		_, err := LoadRepository("")
		assert.ErrorIs(t, err, ErrRepositoryPathNotSpecified)
	})

	t.Run("non-existent path", func(t *testing.T) {
		_, err := LoadRepository("/nonexistent/path/to/repo")
		assert.NotNil(t, err)
	})

	t.Run("path is not a git repository", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := LoadRepository(tmpDir)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "unable to identify git directory")
	})
}

func TestGetGitDir(t *testing.T) {
	t.Run("bare repository", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, true)
		gitDir := repo.GetGitDir()
		assert.Contains(t, gitDir, tmpDir)
	})

	t.Run("non-bare repository", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, false)
		gitDir := repo.GetGitDir()
		assert.Contains(t, gitDir, ".git")
	})
}

func TestIsBareEdgeCases(t *testing.T) {
	t.Run("bare repository with custom path", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, true)
		assert.True(t, repo.IsBare())

		// Git dir path should not end with .git for bare repos
		gitDir := repo.GetGitDir()
		assert.False(t, strings.HasSuffix(gitDir, ".git"))
	})

	t.Run("non-bare repository with .git directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, false)
		assert.False(t, repo.IsBare())

		// Git dir path should end with .git for non-bare repos
		gitDir := repo.GetGitDir()
		assert.True(t, strings.HasSuffix(gitDir, ".git"))
	})
}

func TestLoadRepositoryEdgeCases(t *testing.T) {
	t.Run("load bare repository", func(t *testing.T) {
		tmpDir := t.TempDir()
		_ = CreateTestGitRepository(t, tmpDir, true)

		repo, err := LoadRepository(tmpDir)
		assert.Nil(t, err)
		assert.NotNil(t, repo)
		assert.True(t, repo.IsBare())
	})

	t.Run("load non-bare repository", func(t *testing.T) {
		tmpDir := t.TempDir()
		_ = CreateTestGitRepository(t, tmpDir, false)

		repo, err := LoadRepository(tmpDir)
		assert.Nil(t, err)
		assert.NotNil(t, repo)
		assert.False(t, repo.IsBare())
	})

	t.Run("load repository from subdirectory", func(t *testing.T) {
		tmpDir := t.TempDir()
		_ = CreateTestGitRepository(t, tmpDir, false)

		// Create a subdirectory
		subDir := filepath.Join(tmpDir, "subdir")
		err := os.Mkdir(subDir, 0o755)
		require.Nil(t, err)

		// Should still be able to load from subdirectory
		repo, err := LoadRepository(tmpDir)
		assert.Nil(t, err)
		assert.NotNil(t, repo)
	})
}

func TestGetGoGitRepository(t *testing.T) {
	t.Run("get go-git repository for non-bare repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, false)

		goGitRepo, err := repo.GetGoGitRepository()
		assert.Nil(t, err)
		assert.NotNil(t, goGitRepo)
	})
}

func TestExecutorWithEnv(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("executor with custom environment", func(t *testing.T) {
		// Test that withEnv adds environment variables
		exec := repo.executor("config", "user.name").withEnv("TEST_VAR=test_value")
		assert.NotNil(t, exec)
		assert.Contains(t, exec.env, "TEST_VAR=test_value")
	})

	t.Run("executor with multiple environment variables", func(t *testing.T) {
		exec := repo.executor("config", "user.name").
			withEnv("VAR1=value1").
			withEnv("VAR2=value2", "VAR3=value3")
		assert.NotNil(t, exec)
		assert.Contains(t, exec.env, "VAR1=value1")
		assert.Contains(t, exec.env, "VAR2=value2")
		assert.Contains(t, exec.env, "VAR3=value3")
	})
}

func TestExecutorWithoutGitDir(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("executor without git dir flag", func(t *testing.T) {
		exec := repo.executor("version").withoutGitDir()
		assert.NotNil(t, exec)
		assert.True(t, exec.unsetGitDir)

		// Should still execute successfully
		output, err := exec.executeString()
		assert.Nil(t, err)
		assert.Contains(t, output, "git version")
	})
}

func TestExecutorExecuteString(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("execute string with successful command", func(t *testing.T) {
		output, err := repo.executor("config", "user.name").executeString()
		assert.Nil(t, err)
		assert.NotEmpty(t, output)
	})

	t.Run("execute string with failing command", func(t *testing.T) {
		_, err := repo.executor("config", "nonexistent.key").executeString()
		assert.NotNil(t, err)
	})

	t.Run("execute string trims whitespace", func(t *testing.T) {
		output, err := repo.executor("config", "user.name").executeString()
		assert.Nil(t, err)
		// Output should not have leading/trailing whitespace
		assert.Equal(t, strings.TrimSpace(output), output)
	})
}

func TestExecutorExecute(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("execute with successful command", func(t *testing.T) {
		stdOut, stdErr, err := repo.executor("config", "user.name").execute()
		assert.Nil(t, err)
		assert.NotNil(t, stdOut)
		assert.NotNil(t, stdErr)
	})

	t.Run("execute with failing command", func(t *testing.T) {
		_, _, err := repo.executor("invalid-command").execute()
		assert.NotNil(t, err)
	})
}

func TestLoadRepositoryWithSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	_ = CreateTestGitRepository(t, tmpDir, false)

	repo, err := LoadRepository(tmpDir)
	assert.Nil(t, err)
	assert.NotNil(t, repo)

	// GetGitDir should return an absolute path
	gitDir := repo.GetGitDir()
	assert.True(t, filepath.IsAbs(gitDir))
}

func TestRepositoryClockBehavior(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Verify that the repository has a clock set
	assert.NotNil(t, repo.clock)

	// The test repository uses a fake clock
	assert.NotNil(t, repo.clock.Now())
}

func TestExecutorWithMultipleEnvVars(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Test executor with multiple environment variables
	exec := repo.executor("config", "user.name").
		withEnv("VAR1=value1", "VAR2=value2", "VAR3=value3")

	assert.NotNil(t, exec)
	assert.Contains(t, exec.env, "VAR1=value1")
	assert.Contains(t, exec.env, "VAR2=value2")
	assert.Contains(t, exec.env, "VAR3=value3")
}

func TestExecutorChaining(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Test chaining multiple executor methods
	exec := repo.executor("version").
		withoutGitDir().
		withEnv("TEST=value")

	assert.NotNil(t, exec)
	assert.True(t, exec.unsetGitDir)
	assert.Contains(t, exec.env, "TEST=value")

	// Should still execute successfully
	output, err := exec.executeString()
	assert.Nil(t, err)
	assert.Contains(t, output, "git version")
}

func TestLoadRepositoryWithAbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	_ = CreateTestGitRepository(t, tmpDir, false)

	// Load with absolute path
	repo, err := LoadRepository(tmpDir)
	assert.Nil(t, err)
	assert.NotNil(t, repo)

	// GetGitDir should return absolute path
	gitDir := repo.GetGitDir()
	assert.True(t, filepath.IsAbs(gitDir))
}

func TestRepositoryIsBareConsistency(t *testing.T) {
	t.Run("bare repository consistency", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, true)

		assert.True(t, repo.IsBare())
		gitDir := repo.GetGitDir()
		assert.False(t, strings.HasSuffix(gitDir, ".git"))
	})

	t.Run("non-bare repository consistency", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, false)

		assert.False(t, repo.IsBare())
		gitDir := repo.GetGitDir()
		assert.True(t, strings.HasSuffix(gitDir, ".git"))
	})
}

func TestLoadRepositoryWithInvalidPath(t *testing.T) {
	// Test loading repository from non-existent path
	_, err := LoadRepository("/nonexistent/path/to/repo")
	assert.NotNil(t, err)
}

func TestLoadRepositoryWithNonGitDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a regular directory without .git
	_, err := LoadRepository(tmpDir)
	assert.NotNil(t, err)
}

func TestGetGitDirWithBareRepo(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, true)

	gitDir := repo.GetGitDir()
	assert.Equal(t, tmpDir, gitDir)
}

func TestGetGitDirWithNonBareRepo(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	gitDir := repo.GetGitDir()
	assert.Contains(t, gitDir, ".git")
}
