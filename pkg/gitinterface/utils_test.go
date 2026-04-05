// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package gitinterface

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteRef(t *testing.T) {
	tests := map[string]struct {
		refName    string
		remoteName string
		expected   string
	}{
		"branch ref": {
			refName:    "refs/heads/main",
			remoteName: "origin",
			expected:   "refs/remotes/origin/main",
		},
		"branch ref with path": {
			refName:    "refs/heads/feature/test",
			remoteName: "origin",
			expected:   "refs/remotes/origin/feature/test",
		},
		"tag ref": {
			refName:    "refs/tags/v1.0.0",
			remoteName: "origin",
			expected:   "refs/tags/v1.0.0",
		},
		"custom ref": {
			refName:    "refs/custom/path",
			remoteName: "origin",
			expected:   "refs/remotes/origin/custom/path",
		},
		"different remote name": {
			refName:    "refs/heads/main",
			remoteName: "upstream",
			expected:   "refs/remotes/upstream/main",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := RemoteRef(test.refName, test.remoteName)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestResetDueToError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	refName := "refs/heads/main"
	treeBuilder := NewTreeBuilder(repo)

	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	require.Nil(t, err)

	firstCommitID, err := repo.Commit(emptyTreeID, refName, "Initial commit\n", false)
	require.Nil(t, err)

	_, err = repo.Commit(emptyTreeID, refName, "Second commit\n", false)
	require.Nil(t, err)

	// Ref is at secondCommitID, reset to firstCommitID due to error
	originalErr := assert.AnError
	err = repo.ResetDueToError(originalErr, refName, firstCommitID)
	assert.ErrorIs(t, err, originalErr)

	// Verify ref was reset
	refTip, err := repo.GetReference(refName)
	require.Nil(t, err)
	assert.Equal(t, firstCommitID, refTip)
}

func TestRemoteRefEdgeCases(t *testing.T) {
	tests := map[string]struct {
		refName    string
		remoteName string
		expected   string
	}{
		"branch with nested path": {
			refName:    "refs/heads/feature/nested/path",
			remoteName: "origin",
			expected:   "refs/remotes/origin/feature/nested/path",
		},
		"tag with nested path": {
			refName:    "refs/tags/v1.0.0/beta",
			remoteName: "origin",
			expected:   "refs/tags/v1.0.0/beta",
		},
		"custom ref with nested path": {
			refName:    "refs/custom/namespace/ref",
			remoteName: "upstream",
			expected:   "refs/remotes/upstream/custom/namespace/ref",
		},
		"empty remote name": {
			refName:    "refs/heads/main",
			remoteName: "",
			expected:   "refs/remotes/main",
		},
		"refs prefix only": {
			refName:    "refs/",
			remoteName: "origin",
			expected:   "refs/remotes/origin",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := RemoteRef(test.refName, test.remoteName)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestResetDueToErrorEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	refName := "refs/heads/main"
	treeBuilder := NewTreeBuilder(repo)

	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	require.Nil(t, err)

	firstCommitID, err := repo.Commit(emptyTreeID, refName, "Initial commit\n", false)
	require.Nil(t, err)

	_, err = repo.Commit(emptyTreeID, refName, "Second commit\n", false)
	require.Nil(t, err)

	thirdCommitID, err := repo.Commit(emptyTreeID, refName, "Third commit\n", false)
	require.Nil(t, err)

	// Reset to first commit due to error
	originalErr := assert.AnError
	err = repo.ResetDueToError(originalErr, refName, firstCommitID)
	assert.ErrorIs(t, err, originalErr)

	// Verify ref was reset to first commit
	refTip, err := repo.GetReference(refName)
	require.Nil(t, err)
	assert.Equal(t, firstCommitID, refTip)
	assert.NotEqual(t, thirdCommitID, refTip)
}

func TestIsNiceGitVersion(t *testing.T) {
	// This test just ensures the function runs without error
	// The actual version check depends on the system's Git version
	t.Run("check git version", func(t *testing.T) {
		isNice, err := isNiceGitVersion()
		assert.Nil(t, err)
		// We can't assert the value since it depends on the system
		// but we can ensure it returns a boolean
		assert.IsType(t, true, isNice)
	})
}

func TestTestNameToRefName(t *testing.T) {
	tests := []struct {
		name     string
		testName string
		expected string
	}{
		{
			name:     "simple test name",
			testName: "test",
			expected: "refs/heads/test",
		},
		{
			name:     "test name with spaces",
			testName: "test with spaces",
			expected: "refs/heads/test__with__spaces",
		},
		{
			name:     "test name with multiple spaces",
			testName: "test  with  multiple  spaces",
			expected: "refs/heads/test____with____multiple____spaces",
		},
		{
			name:     "test name with special characters",
			testName: "test-name_123",
			expected: "refs/heads/test-name_123",
		},
		{
			name:     "empty test name",
			testName: "",
			expected: "refs/heads/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testNameToRefName(tt.testName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoteRefWithGittufRefs(t *testing.T) {
	tests := []struct {
		name       string
		refName    string
		remoteName string
		expected   string
	}{
		{
			name:       "gittuf policy ref",
			refName:    "refs/gittuf/policy",
			remoteName: "origin",
			expected:   "refs/remotes/origin/gittuf/policy",
		},
		{
			name:       "gittuf rsl ref",
			refName:    "refs/gittuf/reference-state-log",
			remoteName: "origin",
			expected:   "refs/remotes/origin/gittuf/reference-state-log",
		},
		{
			name:       "gittuf attestations ref",
			refName:    "refs/gittuf/attestations",
			remoteName: "upstream",
			expected:   "refs/remotes/upstream/gittuf/attestations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoteRef(tt.refName, tt.remoteName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResetDueToErrorWithInvalidRef(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Try to reset a non-existent ref
	originalErr := assert.AnError
	err := repo.ResetDueToError(originalErr, "refs/heads/nonexistent", ZeroHash)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, originalErr)
}

func TestRemoteRefWithMultipleRemotes(t *testing.T) {
	remotes := []string{"origin", "upstream", "fork", "mirror"}
	refName := "refs/heads/main"

	for _, remote := range remotes {
		t.Run(remote, func(t *testing.T) {
			result := RemoteRef(refName, remote)
			expected := "refs/remotes/" + remote + "/main"
			assert.Equal(t, expected, result)
		})
	}
}

func TestRemoteRefAllPrefixes(t *testing.T) {
	tests := []struct {
		name       string
		refName    string
		remoteName string
		expected   string
	}{
		{
			name:       "branch ref",
			refName:    "refs/heads/main",
			remoteName: "origin",
			expected:   "refs/remotes/origin/main",
		},
		{
			name:       "branch ref with path",
			refName:    "refs/heads/feature/test",
			remoteName: "origin",
			expected:   "refs/remotes/origin/feature/test",
		},
		{
			name:       "tag ref stays as tag",
			refName:    "refs/tags/v1.0.0",
			remoteName: "origin",
			expected:   "refs/tags/v1.0.0",
		},
		{
			name:       "tag ref with path",
			refName:    "refs/tags/release/v1.0",
			remoteName: "origin",
			expected:   "refs/tags/release/v1.0",
		},
		{
			name:       "custom ref",
			refName:    "refs/custom/path",
			remoteName: "origin",
			expected:   "refs/remotes/origin/custom/path",
		},
		{
			name:       "gittuf ref",
			refName:    "refs/gittuf/policy",
			remoteName: "upstream",
			expected:   "refs/remotes/upstream/gittuf/policy",
		},
		{
			name:       "different remote",
			refName:    "refs/heads/main",
			remoteName: "fork",
			expected:   "refs/remotes/fork/main",
		},
		{
			name:       "empty remote name",
			refName:    "refs/heads/main",
			remoteName: "",
			expected:   "refs/remotes/main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoteRef(tt.refName, tt.remoteName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResetDueToErrorPreservesOriginalError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	treeBuilder := NewTreeBuilder(repo)
	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	require.Nil(t, err)

	commit1, err := repo.Commit(emptyTreeID, "refs/heads/main", "First\n", false)
	require.Nil(t, err)

	commit2, err := repo.Commit(emptyTreeID, "refs/heads/main", "Second\n", false)
	require.Nil(t, err)

	// Create a custom error
	originalErr := fmt.Errorf("custom error occurred")

	// Reset due to error
	err = repo.ResetDueToError(originalErr, "refs/heads/main", commit1)

	// Should return the original error
	assert.ErrorIs(t, err, originalErr)

	// Ref should be reset
	refTip, err := repo.GetReference("refs/heads/main")
	require.Nil(t, err)
	assert.Equal(t, commit1, refTip)
	assert.NotEqual(t, commit2, refTip)
}

func TestTestNameToRefNameVariations(t *testing.T) {
	tests := []struct {
		testName string
		expected string
	}{
		{"simple", "refs/heads/simple"},
		{"with spaces", "refs/heads/with__spaces"},
		{"multiple  spaces", "refs/heads/multiple____spaces"},
		{"test/with/slashes", "refs/heads/test/with/slashes"},
		{"test-with-dashes", "refs/heads/test-with-dashes"},
		{"test_with_underscores", "refs/heads/test_with_underscores"},
		{"MixedCase", "refs/heads/MixedCase"},
		{"", "refs/heads/"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := testNameToRefName(tt.testName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
