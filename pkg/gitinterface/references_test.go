// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package gitinterface

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetReference(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	refName := "refs/heads/main"
	treeBuilder := NewTreeBuilder(repo)

	// Write empty tree
	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	if err != nil {
		t.Fatal(err)
	}

	commitID, err := repo.Commit(emptyTreeID, refName, "Initial commit\n", false)
	require.Nil(t, err)

	refTip, err := repo.GetReference(refName)
	assert.Nil(t, err)
	assert.Equal(t, commitID, refTip)
}

func TestSetReference(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	refName := "refs/heads/main"
	treeBuilder := NewTreeBuilder(repo)

	// Write empty tree
	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	if err != nil {
		t.Fatal(err)
	}

	firstCommitID, err := repo.Commit(emptyTreeID, refName, "Initial commit\n", false)
	require.Nil(t, err)

	// Create second commit with tree
	secondCommitID, err := repo.Commit(emptyTreeID, refName, "Add README\n", false)
	require.Nil(t, err)

	refTip, err := repo.GetReference(refName)
	require.Nil(t, err)
	require.Equal(t, secondCommitID, refTip)

	err = repo.SetReference(refName, firstCommitID)
	assert.Nil(t, err)

	refTip, err = repo.GetReference(refName)
	require.Nil(t, err)
	assert.Equal(t, firstCommitID, refTip)
}

func TestCheckAndSetReference(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	refName := "refs/heads/main"
	treeBuilder := NewTreeBuilder(repo)

	// Write empty tree
	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	if err != nil {
		t.Fatal(err)
	}

	firstCommitID, err := repo.Commit(emptyTreeID, refName, "Initial commit\n", false)
	require.Nil(t, err)

	// Create second commit with tree
	secondCommitID, err := repo.Commit(emptyTreeID, refName, "Add README\n", false)
	require.Nil(t, err)

	refTip, err := repo.GetReference(refName)
	require.Nil(t, err)
	require.Equal(t, secondCommitID, refTip)

	err = repo.CheckAndSetReference(refName, firstCommitID, secondCommitID)
	assert.Nil(t, err)

	refTip, err = repo.GetReference(refName)
	require.Nil(t, err)
	assert.Equal(t, firstCommitID, refTip)
}

func TestGetSymbolicReferenceTarget(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	refName := "refs/heads/main"
	treeBuilder := NewTreeBuilder(repo)

	// Write empty tree
	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.Commit(emptyTreeID, refName, "Initial commit\n", false)
	require.Nil(t, err)

	// HEAD must be set to the main branch -> this is handled by git init
	head, err := repo.GetSymbolicReferenceTarget("HEAD")
	assert.Nil(t, err)
	assert.Equal(t, refName, head)
}

func TestSetSymbolicReference(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	refName := "refs/heads/not-main" // we want to ensure it's set to something other than the default main
	treeBuilder := NewTreeBuilder(repo)

	// Write empty tree
	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.Commit(emptyTreeID, refName, "Initial commit\n", false)
	require.Nil(t, err)

	head, err := repo.GetSymbolicReferenceTarget("HEAD")
	require.Nil(t, err)
	assert.Equal(t, "refs/heads/main", head)

	err = repo.SetSymbolicReference("HEAD", refName)
	assert.Nil(t, err)

	head, err = repo.GetSymbolicReferenceTarget("HEAD")
	require.Nil(t, err)
	assert.Equal(t, refName, head) // not main anymore
}

func TestRepositoryRefSpec(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	shortRefName := "master"
	qualifiedRefName := "refs/heads/master"
	qualifiedRemoteRefName := "refs/remotes/origin/master"

	treeBuilder := NewTreeBuilder(repo)
	emptyTreeHash, err := treeBuilder.WriteTreeFromEntries(nil)
	if err != nil {
		t.Fatal(err)
	}

	commitID, err := repo.Commit(emptyTreeHash, qualifiedRefName, "Test Commit", false)
	if err != nil {
		t.Fatal(err)
	}
	refHash, err := repo.GetReference(qualifiedRefName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, commitID, refHash, "unexpected value configuring test repo")

	tests := map[string]struct {
		repo            *Repository
		refName         string
		remoteName      string
		fastForwardOnly bool
		expectedRefSpec string
		expectedError   error
	}{
		"standard branch, not fast forward only, no remote": {
			refName:         "refs/heads/main",
			expectedRefSpec: "+refs/heads/main:refs/heads/main",
		},
		"standard branch, fast forward only, no remote": {
			refName:         "refs/heads/main",
			fastForwardOnly: true,
			expectedRefSpec: "refs/heads/main:refs/heads/main",
		},
		"standard branch, not fast forward only, remote": {
			refName:         "refs/heads/main",
			remoteName:      "origin",
			expectedRefSpec: "+refs/heads/main:refs/remotes/origin/main",
		},
		"standard branch, fast forward only, remote": {
			refName:         "refs/heads/main",
			remoteName:      "origin",
			fastForwardOnly: true,
			expectedRefSpec: "refs/heads/main:refs/remotes/origin/main",
		},
		"non-standard branch, not fast forward only, no remote": {
			refName:         "refs/heads/foo/bar",
			expectedRefSpec: "+refs/heads/foo/bar:refs/heads/foo/bar",
		},
		"non-standard branch, fast forward only, no remote": {
			refName:         "refs/heads/foo/bar",
			fastForwardOnly: true,
			expectedRefSpec: "refs/heads/foo/bar:refs/heads/foo/bar",
		},
		"non-standard branch, not fast forward only, remote": {
			refName:         "refs/heads/foo/bar",
			remoteName:      "origin",
			expectedRefSpec: "+refs/heads/foo/bar:refs/remotes/origin/foo/bar",
		},
		"non-standard branch, fast forward only, remote": {
			refName:         "refs/heads/foo/bar",
			remoteName:      "origin",
			fastForwardOnly: true,
			expectedRefSpec: "refs/heads/foo/bar:refs/remotes/origin/foo/bar",
		},
		"short branch, not fast forward only, no remote": {
			refName:         shortRefName,
			repo:            repo,
			expectedRefSpec: fmt.Sprintf("+%s:%s", qualifiedRefName, qualifiedRefName),
		},
		"short branch, fast forward only, no remote": {
			refName:         shortRefName,
			repo:            repo,
			fastForwardOnly: true,
			expectedRefSpec: fmt.Sprintf("%s:%s", qualifiedRefName, qualifiedRefName),
		},
		"short branch, not fast forward only, remote": {
			refName:         shortRefName,
			repo:            repo,
			remoteName:      "origin",
			expectedRefSpec: fmt.Sprintf("+%s:%s", qualifiedRefName, qualifiedRemoteRefName),
		},
		"short branch, fast forward only, remote": {
			refName:         shortRefName,
			repo:            repo,
			fastForwardOnly: true,
			remoteName:      "origin",
			expectedRefSpec: fmt.Sprintf("%s:%s", qualifiedRefName, qualifiedRemoteRefName),
		},
		"custom namespace, not fast forward only, no remote": {
			refName:         "refs/foo/bar",
			expectedRefSpec: "+refs/foo/bar:refs/foo/bar",
		},
		"custom namespace, fast forward only, no remote": {
			refName:         "refs/foo/bar",
			fastForwardOnly: true,
			expectedRefSpec: "refs/foo/bar:refs/foo/bar",
		},
		"custom namespace, not fast forward only, remote": {
			refName:         "refs/foo/bar",
			remoteName:      "origin",
			expectedRefSpec: "+refs/foo/bar:refs/remotes/origin/foo/bar",
		},
		"custom namespace, fast forward only, remote": {
			refName:         "refs/foo/bar",
			remoteName:      "origin",
			fastForwardOnly: true,
			expectedRefSpec: "refs/foo/bar:refs/remotes/origin/foo/bar",
		},
		"tag, not fast forward only, no remote": {
			refName:         "refs/tags/v1.0.0",
			fastForwardOnly: false,
			expectedRefSpec: "refs/tags/v1.0.0:refs/tags/v1.0.0",
		},
		"tag, fast forward only, no remote": {
			refName:         "refs/tags/v1.0.0",
			fastForwardOnly: true,
			expectedRefSpec: "refs/tags/v1.0.0:refs/tags/v1.0.0",
		},
		"tag, not fast forward only, remote": {
			refName:         "refs/tags/v1.0.0",
			remoteName:      "origin",
			fastForwardOnly: false,
			expectedRefSpec: "refs/tags/v1.0.0:refs/tags/v1.0.0",
		},
		"tag, fast forward only, remote": {
			refName:         "refs/tags/v1.0.0",
			remoteName:      "origin",
			fastForwardOnly: true,
			expectedRefSpec: "refs/tags/v1.0.0:refs/tags/v1.0.0",
		},
	}

	for name, test := range tests {
		refSpec, err := test.repo.RefSpec(test.refName, test.remoteName, test.fastForwardOnly)
		assert.ErrorIs(t, err, test.expectedError, fmt.Sprintf("unexpected error in test '%s'", name))
		assert.Equal(t, test.expectedRefSpec, refSpec, fmt.Sprintf("unexpected refspec returned in test '%s'", name))
	}
}

func TestBranchReferenceName(t *testing.T) {
	tests := map[string]struct {
		branchName            string
		expectedReferenceName string
	}{
		"short name": {
			branchName:            "main",
			expectedReferenceName: "refs/heads/main",
		},
		"reference name": {
			branchName:            "refs/heads/main",
			expectedReferenceName: "refs/heads/main",
		},
	}

	for name, test := range tests {
		referenceName := BranchReferenceName(test.branchName)
		assert.Equal(t, test.expectedReferenceName, referenceName, fmt.Sprintf("unexpected branch reference received in test '%s'", name))
	}
}

func TestTagReferenceName(t *testing.T) {
	tests := map[string]struct {
		tagName               string
		expectedReferenceName string
	}{
		"short name": {
			tagName:               "v1",
			expectedReferenceName: "refs/tags/v1",
		},
		"reference name": {
			tagName:               "refs/tags/v1",
			expectedReferenceName: "refs/tags/v1",
		},
	}

	for name, test := range tests {
		referenceName := TagReferenceName(test.tagName)
		assert.Equal(t, test.expectedReferenceName, referenceName, fmt.Sprintf("unexpected tag reference received in test '%s'", name))
	}
}

func TestDeleteReference(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	refName := "refs/heads/main"
	treeBuilder := NewTreeBuilder(repo)

	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	if err != nil {
		t.Fatal(err)
	}

	commitID, err := repo.Commit(emptyTreeID, refName, "Initial commit\n", false)
	require.Nil(t, err)

	refTip, err := repo.GetReference(refName)
	require.Nil(t, err)
	require.Equal(t, commitID, refTip)

	err = repo.DeleteReference(refName)
	assert.Nil(t, err)

	_, err = repo.GetReference(refName)
	assert.ErrorIs(t, err, ErrReferenceNotFound)
}

func TestRemoteReferenceName(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"adds prefix if missing": {
			input:    "origin/main",
			expected: "refs/remotes/origin/main",
		},
		"keeps prefix if already present": {
			input:    "refs/remotes/origin/main",
			expected: "refs/remotes/origin/main",
		},
		"empty input returns prefix only": {
			input:    "",
			expected: "refs/remotes/",
		},
		"exact prefix is preserved": {
			input:    "refs/remotes/",
			expected: "refs/remotes/",
		},
	}

	for name, test := range tests {
		referenceName := RemoteReferenceName(test.input)
		assert.Equal(t, test.expected, referenceName, fmt.Sprintf("unexpected remote reference for input %s", name))
	}
}

func TestCustomReferenceName(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"adds prefix if missing": {
			input:    "custom/ref",
			expected: "refs/custom/ref",
		},
		"keeps prefix if already present": {
			input:    "refs/custom/ref",
			expected: "refs/custom/ref",
		},
		"simple name": {
			input:    "myref",
			expected: "refs/myref",
		},
	}

	for name, test := range tests {
		referenceName := CustomReferenceName(test.input)
		assert.Equal(t, test.expected, referenceName, fmt.Sprintf("unexpected custom reference for input %s", name))
	}
}

func TestAbsoluteReference(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	refName := "refs/heads/main"
	treeBuilder := NewTreeBuilder(repo)

	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	require.Nil(t, err)

	_, err = repo.Commit(emptyTreeID, refName, "Initial commit\n", false)
	require.Nil(t, err)

	tests := map[string]struct {
		input    string
		expected string
		hasError bool
	}{
		"short branch name": {
			input:    "main",
			expected: "refs/heads/main",
		},
		"full branch reference": {
			input:    "refs/heads/main",
			expected: "refs/heads/main",
		},
		"HEAD symbolic ref": {
			input:    "HEAD",
			expected: "refs/heads/main",
		},
		"non-existent ref": {
			input:    "nonexistent",
			hasError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := repo.AbsoluteReference(test.input)
			if test.hasError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestGetReferenceNotFound(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	_, err := repo.GetReference("refs/heads/nonexistent")
	assert.ErrorIs(t, err, ErrReferenceNotFound)
}

func TestCheckAndSetReferenceError(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	refName := "refs/heads/main"
	treeBuilder := NewTreeBuilder(repo)

	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	require.Nil(t, err)

	firstCommitID, err := repo.Commit(emptyTreeID, refName, "Initial commit\n", false)
	require.Nil(t, err)

	secondCommitID, err := repo.Commit(emptyTreeID, refName, "Second commit\n", false)
	require.Nil(t, err)

	// Try to set ref to firstCommitID but expect secondCommitID - should fail
	err = repo.CheckAndSetReference(refName, firstCommitID, firstCommitID)
	assert.NotNil(t, err)

	// Ref should still be at secondCommitID
	refTip, err := repo.GetReference(refName)
	require.Nil(t, err)
	assert.Equal(t, secondCommitID, refTip)
}

func TestSetReferenceNewRef(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	treeBuilder := NewTreeBuilder(repo)
	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	require.Nil(t, err)

	commitID, err := repo.Commit(emptyTreeID, "refs/heads/main", "Initial commit\n", false)
	require.Nil(t, err)

	// Set a new reference
	newRefName := "refs/heads/feature"
	err = repo.SetReference(newRefName, commitID)
	assert.Nil(t, err)

	// Verify the reference was set
	refTip, err := repo.GetReference(newRefName)
	assert.Nil(t, err)
	assert.Equal(t, commitID, refTip)
}

func TestDeleteReferenceNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	// Try to delete non-existent reference - should not error
	err := repo.DeleteReference("refs/heads/nonexistent")
	// Git doesn't error when deleting non-existent refs
	assert.Nil(t, err)
}

func TestSetSymbolicReferenceNewSymRef(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	refName := "refs/heads/main"
	treeBuilder := NewTreeBuilder(repo)

	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	require.Nil(t, err)

	_, err = repo.Commit(emptyTreeID, refName, "Initial commit\n", false)
	require.Nil(t, err)

	// Create a new symbolic reference
	symRefName := "refs/heads/current"
	err = repo.SetSymbolicReference(symRefName, refName)
	assert.Nil(t, err)

	// Verify it points to the right target
	target, err := repo.GetSymbolicReferenceTarget(symRefName)
	assert.Nil(t, err)
	assert.Equal(t, refName, target)
}

func TestAbsoluteReferenceEdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	refName := "refs/heads/main"
	treeBuilder := NewTreeBuilder(repo)

	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	require.Nil(t, err)

	_, err = repo.Commit(emptyTreeID, refName, "Initial commit\n", false)
	require.Nil(t, err)

	// Create a tag
	tagName := "refs/tags/v1.0.0"
	commitID, err := repo.GetReference(refName)
	require.Nil(t, err)
	err = repo.SetReference(tagName, commitID)
	require.Nil(t, err)

	// Test with tag short name
	result, err := repo.AbsoluteReference("v1.0.0")
	assert.Nil(t, err)
	assert.Equal(t, tagName, result)

	// Test with already absolute reference
	result, err = repo.AbsoluteReference(refName)
	assert.Nil(t, err)
	assert.Equal(t, refName, result)
}

func TestRefSpecEdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	refName := "refs/heads/main"
	treeBuilder := NewTreeBuilder(repo)

	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	require.Nil(t, err)

	_, err = repo.Commit(emptyTreeID, refName, "Initial commit\n", false)
	require.Nil(t, err)

	// Test with empty remote name
	refSpec, err := repo.RefSpec(refName, "", false)
	assert.Nil(t, err)
	assert.Equal(t, "+refs/heads/main:refs/heads/main", refSpec)

	// Test with remote name
	refSpec, err = repo.RefSpec(refName, "origin", false)
	assert.Nil(t, err)
	assert.Equal(t, "+refs/heads/main:refs/remotes/origin/main", refSpec)

	// Test with fast-forward only
	refSpec, err = repo.RefSpec(refName, "origin", true)
	assert.Nil(t, err)
	assert.Equal(t, "refs/heads/main:refs/remotes/origin/main", refSpec)
}

func TestBranchReferenceNameEdgeCases(t *testing.T) {
	// Test with nested branch name
	result := BranchReferenceName("feature/test")
	assert.Equal(t, "refs/heads/feature/test", result)

	// Test with already prefixed name
	result = BranchReferenceName("refs/heads/main")
	assert.Equal(t, "refs/heads/main", result)

	// Test with empty string
	result = BranchReferenceName("")
	assert.Equal(t, "refs/heads/", result)
}

func TestTagReferenceNameEdgeCases(t *testing.T) {
	// Test with version tag
	result := TagReferenceName("v1.0.0")
	assert.Equal(t, "refs/tags/v1.0.0", result)

	// Test with already prefixed name
	result = TagReferenceName("refs/tags/v1.0.0")
	assert.Equal(t, "refs/tags/v1.0.0", result)

	// Test with empty string
	result = TagReferenceName("")
	assert.Equal(t, "refs/tags/", result)
}

func TestCustomReferenceNameEdgeCases(t *testing.T) {
	// Test with nested custom ref
	result := CustomReferenceName("custom/namespace/ref")
	assert.Equal(t, "refs/custom/namespace/ref", result)

	// Test with already prefixed name
	result = CustomReferenceName("refs/custom/ref")
	assert.Equal(t, "refs/custom/ref", result)

	// Test with empty string
	result = CustomReferenceName("")
	assert.Equal(t, "refs/", result)
}

func TestReferenceOperationsComprehensive(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	treeBuilder := NewTreeBuilder(repo)
	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	require.Nil(t, err)

	// Create multiple commits and refs
	refs := []string{
		"refs/heads/main",
		"refs/heads/develop",
		"refs/heads/feature/test",
		"refs/tags/v1.0.0",
		"refs/custom/myref",
	}

	commitIDs := make(map[string]Hash)
	for _, refName := range refs {
		commitID, err := repo.Commit(emptyTreeID, refName, fmt.Sprintf("Commit for %s\n", refName), false)
		assert.Nil(t, err)
		commitIDs[refName] = commitID

		// Test GetReference
		retrievedID, err := repo.GetReference(refName)
		assert.Nil(t, err)
		assert.Equal(t, commitID, retrievedID)

		// Test SetReference
		newCommitID, err := repo.Commit(emptyTreeID, refName, fmt.Sprintf("New commit for %s\n", refName), false)
		assert.Nil(t, err)

		err = repo.SetReference(refName, commitID)
		assert.Nil(t, err)

		retrievedID, err = repo.GetReference(refName)
		assert.Nil(t, err)
		assert.Equal(t, commitID, retrievedID)

		// Reset to new commit
		err = repo.SetReference(refName, newCommitID)
		assert.Nil(t, err)
	}
}

func TestAbsoluteReferenceComprehensive(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	treeBuilder := NewTreeBuilder(repo)
	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	require.Nil(t, err)

	// Create a commit
	_, err = repo.Commit(emptyTreeID, "refs/heads/main", "Test\n", false)
	require.Nil(t, err)

	t.Run("already absolute branch ref", func(t *testing.T) {
		absRef, err := repo.AbsoluteReference("refs/heads/main")
		assert.Nil(t, err)
		assert.Equal(t, "refs/heads/main", absRef)
	})

	t.Run("short branch name", func(t *testing.T) {
		absRef, err := repo.AbsoluteReference("main")
		assert.Nil(t, err)
		assert.Equal(t, "refs/heads/main", absRef)
	})

	t.Run("already absolute tag ref", func(t *testing.T) {
		absRef, err := repo.AbsoluteReference("refs/tags/v1.0")
		assert.Nil(t, err)
		assert.Equal(t, "refs/tags/v1.0", absRef)
	})

	t.Run("short tag name without tag existing", func(t *testing.T) {
		// Short tag names need the tag to exist first
		// This test verifies the error case when tag doesn't exist
		_, err := repo.AbsoluteReference("v1.0")
		assert.ErrorIs(t, err, ErrReferenceNotFound)
	})

	t.Run("non-existent ref", func(t *testing.T) {
		_, err := repo.AbsoluteReference("nonexistent")
		assert.ErrorIs(t, err, ErrReferenceNotFound)
	})

	t.Run("HEAD symbolic ref", func(t *testing.T) {
		absRef, err := repo.AbsoluteReference("HEAD")
		assert.Nil(t, err)
		assert.Equal(t, "refs/heads/main", absRef)
	})
}

func TestRefSpecComprehensive(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	treeBuilder := NewTreeBuilder(repo)
	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	require.Nil(t, err)

	_, err = repo.Commit(emptyTreeID, "refs/heads/main", "Test\n", false)
	require.Nil(t, err)

	t.Run("branch with fast-forward only", func(t *testing.T) {
		refSpec, err := repo.RefSpec("main", "origin", true)
		assert.Nil(t, err)
		assert.Equal(t, "refs/heads/main:refs/remotes/origin/main", refSpec)
	})

	t.Run("branch without fast-forward only", func(t *testing.T) {
		refSpec, err := repo.RefSpec("main", "origin", false)
		assert.Nil(t, err)
		assert.Equal(t, "+refs/heads/main:refs/remotes/origin/main", refSpec)
	})

	t.Run("absolute ref with remote", func(t *testing.T) {
		refSpec, err := repo.RefSpec("refs/heads/main", "upstream", true)
		assert.Nil(t, err)
		assert.Equal(t, "refs/heads/main:refs/remotes/upstream/main", refSpec)
	})

	t.Run("ref without remote name", func(t *testing.T) {
		refSpec, err := repo.RefSpec("refs/heads/main", "", true)
		assert.Nil(t, err)
		assert.Equal(t, "refs/heads/main:refs/heads/main", refSpec)
	})

	t.Run("non-existent ref", func(t *testing.T) {
		_, err := repo.RefSpec("nonexistent", "origin", true)
		assert.NotNil(t, err)
	})
}

func TestReferenceNameHelpers(t *testing.T) {
	t.Run("CustomReferenceName", func(t *testing.T) {
		assert.Equal(t, "refs/custom", CustomReferenceName("custom"))
		assert.Equal(t, "refs/custom/path", CustomReferenceName("refs/custom/path"))
	})

	t.Run("TagReferenceName", func(t *testing.T) {
		assert.Equal(t, "refs/tags/v1.0", TagReferenceName("v1.0"))
		assert.Equal(t, "refs/tags/v1.0", TagReferenceName("refs/tags/v1.0"))
	})

	t.Run("BranchReferenceName", func(t *testing.T) {
		assert.Equal(t, "refs/heads/main", BranchReferenceName("main"))
		assert.Equal(t, "refs/heads/main", BranchReferenceName("refs/heads/main"))
	})

	t.Run("RemoteReferenceName", func(t *testing.T) {
		assert.Equal(t, "refs/remotes/origin", RemoteReferenceName("origin"))
		assert.Equal(t, "refs/remotes/origin", RemoteReferenceName("refs/remotes/origin"))
	})
}

func TestGetSymbolicReferenceTargetError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Try to get symbolic ref target of non-symbolic ref
	treeBuilder := NewTreeBuilder(repo)
	emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
	require.Nil(t, err)

	_, err = repo.Commit(emptyTreeID, "refs/heads/main", "Test\n", false)
	require.Nil(t, err)

	_, err = repo.GetSymbolicReferenceTarget("refs/heads/main")
	assert.NotNil(t, err)
}

func TestSetSymbolicReferenceError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Try to set symbolic ref with invalid target
	err := repo.SetSymbolicReference("HEAD", "")
	assert.NotNil(t, err)
}
