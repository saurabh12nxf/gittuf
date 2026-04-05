// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package gitinterface

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemote(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	output, err := repo.executor("remote").executeString()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "", output) // no output because there are no remotes

	remoteName := "origin"
	remoteURL := "git@example.com:repo.git"

	// Test AddRemote
	err = repo.AddRemote(remoteName, remoteURL)
	assert.Nil(t, err)

	output, err = repo.executor("remote", "-v").executeString()
	if err != nil {
		t.Fatal(err)
	}

	expectedOutput := fmt.Sprintf("%s\t%s (fetch)\n%s\t%s (push)", remoteName, remoteURL, remoteName, remoteURL)
	assert.Equal(t, expectedOutput, output)

	// Test GetRemoteURL
	returnedRemoteURL, err := repo.GetRemoteURL(remoteName)
	assert.Nil(t, err)
	assert.Equal(t, remoteURL, returnedRemoteURL)

	_, err = repo.GetRemoteURL("does-not-exist")
	assert.ErrorContains(t, err, "No such remote")

	// Test RemoveRemote
	err = repo.RemoveRemote(remoteName)
	assert.Nil(t, err)

	output, err = repo.executor("remote").executeString()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "", output) // no output because there are no remotes
}

func TestAddRemoteError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	remoteName := "origin"
	remoteURL := "git@example.com:repo.git"

	// Add remote
	err := repo.AddRemote(remoteName, remoteURL)
	assert.Nil(t, err)

	// Try to add same remote again - should fail
	err = repo.AddRemote(remoteName, remoteURL)
	assert.NotNil(t, err)
}

func TestRemoveRemoteError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Try to remove non-existent remote
	err := repo.RemoveRemote("nonexistent")
	assert.NotNil(t, err)
}

func TestGetRemoteURLEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("get URL for multiple remotes", func(t *testing.T) {
		// Add multiple remotes
		err := repo.AddRemote("origin", "git@github.com:user/repo1.git")
		assert.Nil(t, err)

		err = repo.AddRemote("upstream", "git@github.com:org/repo2.git")
		assert.Nil(t, err)

		// Get URLs
		originURL, err := repo.GetRemoteURL("origin")
		assert.Nil(t, err)
		assert.Equal(t, "git@github.com:user/repo1.git", originURL)

		upstreamURL, err := repo.GetRemoteURL("upstream")
		assert.Nil(t, err)
		assert.Equal(t, "git@github.com:org/repo2.git", upstreamURL)
	})

	t.Run("get URL with https protocol", func(t *testing.T) {
		err := repo.AddRemote("https-remote", "https://github.com/user/repo.git")
		assert.Nil(t, err)

		url, err := repo.GetRemoteURL("https-remote")
		assert.Nil(t, err)
		assert.Equal(t, "https://github.com/user/repo.git", url)
	})

	t.Run("get URL with file protocol", func(t *testing.T) {
		localPath := t.TempDir()
		err := repo.AddRemote("local", localPath)
		assert.Nil(t, err)

		url, err := repo.GetRemoteURL("local")
		assert.Nil(t, err)
		assert.Contains(t, url, localPath)
	})
}

func TestRemoveRemoteEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("remove one of multiple remotes", func(t *testing.T) {
		// Add multiple remotes
		err := repo.AddRemote("origin", "git@example.com:repo1.git")
		assert.Nil(t, err)

		err = repo.AddRemote("upstream", "git@example.com:repo2.git")
		assert.Nil(t, err)

		// Remove one
		err = repo.RemoveRemote("origin")
		assert.Nil(t, err)

		// Verify origin is gone but upstream remains
		_, err = repo.GetRemoteURL("origin")
		assert.NotNil(t, err)

		url, err := repo.GetRemoteURL("upstream")
		assert.Nil(t, err)
		assert.Equal(t, "git@example.com:repo2.git", url)
	})
}

func TestGetRemoteURL(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("get URL of existing remote", func(t *testing.T) {
		err := repo.AddRemote("test-remote", "https://example.com/repo.git")
		require.Nil(t, err)

		url, err := repo.GetRemoteURL("test-remote")
		assert.Nil(t, err)
		assert.Equal(t, "https://example.com/repo.git", url)
	})

	t.Run("get URL of non-existent remote", func(t *testing.T) {
		_, err := repo.GetRemoteURL("non-existent-remote")
		assert.NotNil(t, err)
	})

	t.Run("get URL after updating remote", func(t *testing.T) {
		err := repo.AddRemote("update-remote", "https://old-url.com/repo.git")
		require.Nil(t, err)

		// Remove and re-add with new URL
		err = repo.RemoveRemote("update-remote")
		require.Nil(t, err)

		err = repo.AddRemote("update-remote", "https://new-url.com/repo.git")
		require.Nil(t, err)

		url, err := repo.GetRemoteURL("update-remote")
		assert.Nil(t, err)
		assert.Equal(t, "https://new-url.com/repo.git", url)
	})
}

func TestAddRemoteComprehensive(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("add remote with HTTPS URL", func(t *testing.T) {
		err := repo.AddRemote("https-remote", "https://github.com/user/repo.git")
		assert.Nil(t, err)

		url, err := repo.GetRemoteURL("https-remote")
		assert.Nil(t, err)
		assert.Equal(t, "https://github.com/user/repo.git", url)
	})

	t.Run("add remote with SSH URL", func(t *testing.T) {
		err := repo.AddRemote("ssh-remote", "git@github.com:user/repo.git")
		assert.Nil(t, err)

		url, err := repo.GetRemoteURL("ssh-remote")
		assert.Nil(t, err)
		assert.Equal(t, "git@github.com:user/repo.git", url)
	})

	t.Run("add remote with file path", func(t *testing.T) {
		err := repo.AddRemote("file-remote", "/path/to/repo.git")
		assert.Nil(t, err)

		url, err := repo.GetRemoteURL("file-remote")
		assert.Nil(t, err)
		assert.Equal(t, "/path/to/repo.git", url)
	})

	t.Run("add remote with special characters in name", func(t *testing.T) {
		err := repo.AddRemote("remote-with-dash", "https://example.com/repo.git")
		assert.Nil(t, err)

		url, err := repo.GetRemoteURL("remote-with-dash")
		assert.Nil(t, err)
		assert.Equal(t, "https://example.com/repo.git", url)
	})
}

func TestRemoveRemoteComprehensive(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("remove existing remote", func(t *testing.T) {
		err := repo.AddRemote("to-remove", "https://example.com/repo.git")
		require.Nil(t, err)

		err = repo.RemoveRemote("to-remove")
		assert.Nil(t, err)

		// Verify it's removed
		_, err = repo.GetRemoteURL("to-remove")
		assert.NotNil(t, err)
	})

	t.Run("remove non-existent remote", func(t *testing.T) {
		err := repo.RemoveRemote("does-not-exist")
		assert.NotNil(t, err)
	})

	t.Run("remove and re-add remote", func(t *testing.T) {
		err := repo.AddRemote("readd-remote", "https://example.com/repo1.git")
		require.Nil(t, err)

		err = repo.RemoveRemote("readd-remote")
		require.Nil(t, err)

		err = repo.AddRemote("readd-remote", "https://example.com/repo2.git")
		assert.Nil(t, err)

		url, err := repo.GetRemoteURL("readd-remote")
		assert.Nil(t, err)
		assert.Equal(t, "https://example.com/repo2.git", url)
	})
}

func TestRemoteOperationsSequence(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Add multiple remotes
	remotes := map[string]string{
		"origin":   "https://github.com/user/repo.git",
		"upstream": "https://github.com/upstream/repo.git",
		"fork":     "https://github.com/fork/repo.git",
	}

	for name, url := range remotes {
		err := repo.AddRemote(name, url)
		require.Nil(t, err)
	}

	// Verify all remotes
	for name, expectedURL := range remotes {
		url, err := repo.GetRemoteURL(name)
		assert.Nil(t, err)
		assert.Equal(t, expectedURL, url)
	}

	// Remove one remote
	err := repo.RemoveRemote("fork")
	require.Nil(t, err)

	// Verify fork is removed
	_, err = repo.GetRemoteURL("fork")
	assert.NotNil(t, err)

	// Verify others still exist
	url, err := repo.GetRemoteURL("origin")
	assert.Nil(t, err)
	assert.Equal(t, remotes["origin"], url)
}

func TestCreateRemoteWithSpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Test remote with special characters in URL
	err := repo.CreateRemote("special", "https://user:pass@example.com:8080/repo.git")
	assert.Nil(t, err)

	// Verify remote was created by getting its URL
	url, err := repo.GetRemoteURL("special")
	assert.Nil(t, err)
	assert.Contains(t, url, "example.com")
}

func TestRemoveNonExistentRemote(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Try to remove a remote that doesn't exist
	err := repo.RemoveRemote("nonexistent")
	assert.NotNil(t, err)
}

func TestGetRemoteURLForNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Try to get URL for non-existent remote
	_, err := repo.GetRemoteURL("nonexistent")
	assert.NotNil(t, err)
}
