// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package gitinterface

import (
	"fmt"
	"testing"

	artifacts "github.com/gittuf/gittuf/internal/testartifacts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasObject(t *testing.T) {
	tempDir1 := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir1, true)

	// Create a backup repo to compute Git IDs we test in repo
	tempDir2 := t.TempDir()
	backupRepo := CreateTestGitRepository(t, tempDir2, true)

	blobID, err := backupRepo.WriteBlob([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, backupRepo.HasObject(blobID)) // backup has it
	assert.False(t, repo.HasObject(blobID))      // repo does not

	if _, err := repo.WriteBlob([]byte("hello")); err != nil {
		t.Fatal(err)
	}

	assert.True(t, repo.HasObject(blobID)) // now repo has it too

	backupRepoTreeBuilder := NewTreeBuilder(backupRepo)
	treeID, err := backupRepoTreeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("file", blobID)})
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, backupRepo.HasObject(treeID)) // backup has it
	assert.False(t, repo.HasObject(treeID))      // repo does not

	repoTreeBuilder := NewTreeBuilder(repo)
	if _, err := repoTreeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("file", blobID)}); err != nil {
		t.Fatal(err)
	}

	assert.True(t, repo.HasObject(treeID)) // now repo has it too

	commitID, err := backupRepo.Commit(treeID, "refs/heads/main", "Initial commit\n", false)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, backupRepo.HasObject(commitID)) // backup has it
	assert.False(t, repo.HasObject(commitID))      // repo does not

	if _, err := repo.Commit(treeID, "refs/heads/main", "Initial commit\n", false); err != nil {
		t.Fatal(err)
	}

	// Note: This test passes because we control timestamps in
	// CreateTestGitRepository. So, commit ID in both repos is the same.
	assert.True(t, repo.HasObject(commitID)) // now repo has it too
}

func TestGetObjectType(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	blobID, err := repo.WriteBlob([]byte("gittuf"))
	require.Nil(t, err)

	objType, err := repo.GetObjectType(blobID)
	assert.Nil(t, err)
	assert.Equal(t, BlobObjectType, objType)

	treeBuilder := NewTreeBuilder(repo)
	treeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("foo", blobID)})
	require.Nil(t, err)

	objType, err = repo.GetObjectType(treeID)
	assert.Nil(t, err)
	assert.Equal(t, TreeObjectType, objType)

	commitID, err := repo.Commit(treeID, "refs/heads/main", "Test commit\n", false)
	require.Nil(t, err)

	objType, err = repo.GetObjectType(commitID)
	assert.Nil(t, err)
	assert.Equal(t, CommitObjectType, objType)

	tagID, err := repo.TagUsingSpecificKey(commitID, "test-tag", "Test tag\n", artifacts.GPGKey1Private)
	require.Nil(t, err)

	objType, err = repo.GetObjectType(tagID)
	assert.Nil(t, err)
	assert.Equal(t, TagObjectType, objType)
}

func TestGetObjectSize(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	blobID, err := repo.WriteBlob([]byte("gittuf"))
	require.Nil(t, err)

	objSize, err := repo.GetObjectSize(blobID)
	assert.Nil(t, err)
	assert.Equal(t, uint64(6), objSize)
}

func TestGetObjectTypeError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Try with non-existent object
	_, err := repo.GetObjectType(ZeroHash)
	assert.NotNil(t, err)
}

func TestGetObjectSizeError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Try with non-existent object
	_, err := repo.GetObjectSize(ZeroHash)
	assert.NotNil(t, err)
}

func TestHasObjectWithZeroHash(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// ZeroHash should not exist
	assert.False(t, repo.HasObject(ZeroHash))
}

func TestHasObjectEdgeCases(t *testing.T) {
	tempDir1 := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir1, true)

	t.Run("has blob object", func(t *testing.T) {
		blobID, err := repo.WriteBlob([]byte("test content"))
		require.Nil(t, err)

		assert.True(t, repo.HasObject(blobID))
	})

	t.Run("has tree object", func(t *testing.T) {
		blobID, err := repo.WriteBlob([]byte("test"))
		require.Nil(t, err)

		treeBuilder := NewTreeBuilder(repo)
		treeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("file.txt", blobID),
		})
		require.Nil(t, err)

		assert.True(t, repo.HasObject(treeID))
	})

	t.Run("has commit object", func(t *testing.T) {
		treeBuilder := NewTreeBuilder(repo)
		treeID, err := treeBuilder.WriteTreeFromEntries(nil)
		require.Nil(t, err)

		commitID, err := repo.Commit(treeID, "refs/heads/main", "Test commit\n", false)
		require.Nil(t, err)

		assert.True(t, repo.HasObject(commitID))
	})

	t.Run("does not have non-existent object", func(t *testing.T) {
		nonExistentHash, err := NewHash("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		require.Nil(t, err)

		assert.False(t, repo.HasObject(nonExistentHash))
	})
}

func TestGetObjectSizeEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("size of empty blob", func(t *testing.T) {
		blobID, err := repo.WriteBlob([]byte{})
		require.Nil(t, err)

		size, err := repo.GetObjectSize(blobID)
		assert.Nil(t, err)
		assert.Equal(t, uint64(0), size)
	})

	t.Run("size of small blob", func(t *testing.T) {
		content := []byte("hello")
		blobID, err := repo.WriteBlob(content)
		require.Nil(t, err)

		size, err := repo.GetObjectSize(blobID)
		assert.Nil(t, err)
		assert.Equal(t, uint64(len(content)), size)
	})

	t.Run("size of large blob", func(t *testing.T) {
		content := make([]byte, 10000)
		for i := range content {
			content[i] = byte(i % 256)
		}
		blobID, err := repo.WriteBlob(content)
		require.Nil(t, err)

		size, err := repo.GetObjectSize(blobID)
		assert.Nil(t, err)
		assert.Equal(t, uint64(len(content)), size)
	})
}

func TestGetObjectTypeAllTypes(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("blob type", func(t *testing.T) {
		blobID, err := repo.WriteBlob([]byte("test"))
		require.Nil(t, err)

		objType, err := repo.GetObjectType(blobID)
		assert.Nil(t, err)
		assert.Equal(t, BlobObjectType, objType)
	})

	t.Run("tree type", func(t *testing.T) {
		treeBuilder := NewTreeBuilder(repo)
		treeID, err := treeBuilder.WriteTreeFromEntries(nil)
		require.Nil(t, err)

		objType, err := repo.GetObjectType(treeID)
		assert.Nil(t, err)
		assert.Equal(t, TreeObjectType, objType)
	})

	t.Run("commit type", func(t *testing.T) {
		treeBuilder := NewTreeBuilder(repo)
		treeID, err := treeBuilder.WriteTreeFromEntries(nil)
		require.Nil(t, err)

		commitID, err := repo.Commit(treeID, "refs/heads/main", "Test\n", false)
		require.Nil(t, err)

		objType, err := repo.GetObjectType(commitID)
		assert.Nil(t, err)
		assert.Equal(t, CommitObjectType, objType)
	})

	t.Run("tag type", func(t *testing.T) {
		treeBuilder := NewTreeBuilder(repo)
		treeID, err := treeBuilder.WriteTreeFromEntries(nil)
		require.Nil(t, err)

		commitID, err := repo.Commit(treeID, "refs/heads/main", "Test\n", false)
		require.Nil(t, err)

		tagID, err := repo.TagUsingSpecificKey(commitID, "v1.0", "Release\n", artifacts.SSHED25519Private)
		require.Nil(t, err)

		objType, err := repo.GetObjectType(tagID)
		assert.Nil(t, err)
		assert.Equal(t, TagObjectType, objType)
	})

	t.Run("non-existent object", func(t *testing.T) {
		_, err := repo.GetObjectType(ZeroHash)
		assert.NotNil(t, err)
	})
}

func TestGetObjectSizeVariousSizes(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	sizes := []int{0, 1, 10, 100, 1000, 10000}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			content := make([]byte, size)
			for i := range content {
				content[i] = byte(i % 256)
			}

			blobID, err := repo.WriteBlob(content)
			require.Nil(t, err)

			objSize, err := repo.GetObjectSize(blobID)
			assert.Nil(t, err)
			assert.Equal(t, uint64(size), objSize) //nolint:gosec // Test size comparison
		})
	}
}

func TestHasObjectWithDifferentTypes(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Test with blob
	blobID, err := repo.WriteBlob([]byte("blob"))
	require.Nil(t, err)
	assert.True(t, repo.HasObject(blobID))

	// Test with tree
	treeBuilder := NewTreeBuilder(repo)
	treeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
		NewEntryBlob("file.txt", blobID),
	})
	require.Nil(t, err)
	assert.True(t, repo.HasObject(treeID))

	// Test with commit
	commitID, err := repo.Commit(treeID, "refs/heads/main", "Commit\n", false)
	require.Nil(t, err)
	assert.True(t, repo.HasObject(commitID))

	// Test with non-existent
	fakeHash, _ := NewHash("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.False(t, repo.HasObject(fakeHash))
}

func TestGetObjectSizeForDifferentTypes(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	t.Run("blob size", func(t *testing.T) {
		content := []byte("test content with specific length")
		blobID, err := repo.WriteBlob(content)
		require.Nil(t, err)

		size, err := repo.GetObjectSize(blobID)
		assert.Nil(t, err)
		assert.Equal(t, uint64(len(content)), size)
	})

	t.Run("tree size", func(t *testing.T) {
		emptyTreeID, err := repo.EmptyTree()
		require.Nil(t, err)

		size, err := repo.GetObjectSize(emptyTreeID)
		assert.Nil(t, err)
		assert.GreaterOrEqual(t, size, uint64(0))
	})

	t.Run("commit size", func(t *testing.T) {
		emptyTreeID, err := repo.EmptyTree()
		require.Nil(t, err)

		commitID, err := repo.Commit(emptyTreeID, "refs/heads/main", "Test\n", false)
		require.Nil(t, err)

		size, err := repo.GetObjectSize(commitID)
		assert.Nil(t, err)
		assert.Greater(t, size, uint64(0))
	})

	t.Run("invalid object", func(t *testing.T) {
		invalidHash, _ := NewHash("0000000000000000000000000000000000000001")
		_, err := repo.GetObjectSize(invalidHash)
		assert.NotNil(t, err)
	})
}

func TestHasObjectWithInvalidHash(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Test with a hash that doesn't exist
	invalidHash, _ := NewHash("1234567890123456789012345678901234567890")
	has := repo.HasObject(invalidHash)
	assert.False(t, has)
}
