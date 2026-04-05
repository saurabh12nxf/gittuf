// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package gitinterface

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryEmptyTree(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	hash, err := repo.EmptyTree()
	assert.Nil(t, err)

	// SHA-1 ID used by Git to denote an empty tree
	// $ git hash-object -t tree --stdin < /dev/null
	assert.Equal(t, "4b825dc642cb6eb9a060e54bf8d69288fbee4904", hash.String())
}

func TestGetPathIDInTree(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)
	treeBuilder := NewTreeBuilder(repo)

	blobAID, err := repo.WriteBlob([]byte("a"))
	if err != nil {
		t.Fatal(err)
	}

	blobBID, err := repo.WriteBlob([]byte("b"))
	if err != nil {
		t.Fatal(err)
	}

	emptyTreeID := "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

	t.Run("no items", func(t *testing.T) {
		treeID, err := treeBuilder.WriteTreeFromEntries(nil)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, emptyTreeID, treeID.String())

		pathID, err := repo.GetPathIDInTree("a", treeID)
		assert.ErrorIs(t, err, ErrTreeDoesNotHavePath)
		assert.Nil(t, pathID)
	})

	t.Run("no subdirectories", func(t *testing.T) {
		exhaustiveItems := []TreeEntry{
			NewEntryBlob("a", blobAID),
			NewEntryBlob("b", blobBID),
		}

		treeID, err := treeBuilder.WriteTreeFromEntries(exhaustiveItems)
		if err != nil {
			t.Fatal(err)
		}

		itemID, err := repo.GetPathIDInTree("a", treeID)
		assert.Nil(t, err)
		assert.Equal(t, blobAID, itemID)
	})

	t.Run("one file in root tree, one file in subdirectory", func(t *testing.T) {
		exhaustiveItems := []TreeEntry{
			NewEntryBlob("foo/a", blobAID),
			NewEntryBlob("b", blobBID),
		}

		treeID, err := treeBuilder.WriteTreeFromEntries(exhaustiveItems)
		if err != nil {
			t.Fatal(err)
		}

		itemID, err := repo.GetPathIDInTree("foo/a", treeID)
		assert.Nil(t, err)
		assert.Equal(t, blobAID, itemID)
	})

	t.Run("multiple levels", func(t *testing.T) {
		exhaustiveItems := []TreeEntry{
			NewEntryBlob("foo/bar/foobar/a", blobAID),
			NewEntryBlob("foobar/foo/bar/b", blobBID),
		}

		treeID, err := treeBuilder.WriteTreeFromEntries(exhaustiveItems)
		if err != nil {
			t.Fatal(err)
		}

		// find tree ID for foo/bar/foobar
		expectedItemID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("a", blobAID)})
		if err != nil {
			t.Fatal(err)
		}

		itemID, err := repo.GetPathIDInTree("foo/bar/foobar", treeID)
		assert.Nil(t, err)
		assert.Equal(t, expectedItemID, itemID)

		// find tree ID for foo/bar
		expectedItemID, err = treeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("foobar/a", blobAID)})
		if err != nil {
			t.Fatal(err)
		}

		itemID, err = repo.GetPathIDInTree("foo/bar", treeID)
		assert.Nil(t, err)
		assert.Equal(t, expectedItemID, itemID)

		// find tree ID for foobar/foo
		expectedItemID, err = treeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("bar/b", blobBID)})
		if err != nil {
			t.Fatal(err)
		}

		itemID, err = repo.GetPathIDInTree("foobar/foo", treeID)
		assert.Nil(t, err)
		assert.Equal(t, expectedItemID, itemID)

		itemID, err = repo.GetPathIDInTree("foobar/foo/foobar", treeID)
		assert.ErrorIs(t, err, ErrTreeDoesNotHavePath)
		assert.Nil(t, itemID)
	})
}

func TestGetTreeItems(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)
	treeBuilder := NewTreeBuilder(repo)

	blobAID, err := repo.WriteBlob([]byte("a"))
	if err != nil {
		t.Fatal(err)
	}

	blobBID, err := repo.WriteBlob([]byte("b"))
	if err != nil {
		t.Fatal(err)
	}

	emptyTreeID := "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

	t.Run("no items", func(t *testing.T) {
		treeID, err := treeBuilder.WriteTreeFromEntries(nil)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, emptyTreeID, treeID.String())

		treeItems, err := repo.GetTreeItems(treeID)
		assert.Nil(t, err)
		assert.Nil(t, treeItems)
	})

	t.Run("no subdirectories", func(t *testing.T) {
		exhaustiveItems := []TreeEntry{
			NewEntryBlob("a", blobAID),
			NewEntryBlob("b", blobBID),
		}

		treeID, err := treeBuilder.WriteTreeFromEntries(exhaustiveItems)
		if err != nil {
			t.Fatal(err)
		}

		expectedOutput := map[string]Hash{
			"a": blobAID,
			"b": blobBID,
		}

		treeItems, err := repo.GetTreeItems(treeID)
		assert.Nil(t, err)
		assert.Equal(t, expectedOutput, treeItems)
	})

	t.Run("one file in root tree, one file in subdirectory", func(t *testing.T) {
		exhaustiveItems := []TreeEntry{
			NewEntryBlob("foo/a", blobAID),
			NewEntryBlob("b", blobBID),
		}

		treeID, err := treeBuilder.WriteTreeFromEntries(exhaustiveItems)
		if err != nil {
			t.Fatal(err)
		}

		fooTreeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("a", blobAID)})
		if err != nil {
			t.Fatal(err)
		}

		expectedTreeItems := map[string]Hash{
			"foo": fooTreeID,
			"b":   blobBID,
		}

		treeItems, err := repo.GetTreeItems(treeID)
		assert.Nil(t, err)
		assert.Equal(t, expectedTreeItems, treeItems)
	})

	t.Run("one file in foo tree, one file in bar", func(t *testing.T) {
		exhaustiveItems := []TreeEntry{
			NewEntryBlob("foo/a", blobAID),
			NewEntryBlob("bar/b", blobBID),
		}

		treeID, err := treeBuilder.WriteTreeFromEntries(exhaustiveItems)
		if err != nil {
			t.Fatal(err)
		}

		fooTreeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("a", blobAID)})
		if err != nil {
			t.Fatal(err)
		}

		barTreeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("b", blobBID)})
		if err != nil {
			t.Fatal(err)
		}

		expectedTreeItems := map[string]Hash{
			"foo": fooTreeID,
			"bar": barTreeID,
		}

		treeItems, err := repo.GetTreeItems(treeID)
		assert.Nil(t, err)
		assert.Equal(t, expectedTreeItems, treeItems)
	})
}

func TestGetMergeTree(t *testing.T) {
	t.Run("no conflict", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, false)

		// We meed to change the directory for this test because we `checkout`
		// for older Git versions, modifying the worktree. This chdir ensures
		// that the temporary directory is used as the worktree.
		pwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(pwd) //nolint:errcheck

		emptyBlobID, err := repo.WriteBlob(nil)
		if err != nil {
			t.Fatal(err)
		}

		treeBuilder := NewTreeBuilder(repo)
		emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
		if err != nil {
			t.Fatal(err)
		}

		treeAID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("a", emptyBlobID)})
		if err != nil {
			t.Fatal(err)
		}
		treeBID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("b", emptyBlobID)})
		if err != nil {
			t.Fatal(err)
		}
		combinedTreeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("a", emptyBlobID),
			NewEntryBlob("b", emptyBlobID),
		})
		if err != nil {
			t.Fatal(err)
		}

		mainRef := "refs/heads/main"
		featureRef := "refs/heads/feature"

		// Add commits to the main branch
		baseCommitID, err := repo.Commit(emptyTreeID, mainRef, "Initial commit", false)
		if err != nil {
			t.Fatal(err)
		}
		commitAID, err := repo.Commit(treeAID, mainRef, "Commit A", false)
		if err != nil {
			t.Fatal(err)
		}

		// Add commits to the feature branch
		if err := repo.SetReference(featureRef, baseCommitID); err != nil {
			t.Fatal(err)
		}
		commitBID, err := repo.Commit(treeBID, featureRef, "Commit B", false)
		if err != nil {
			t.Fatal(err)
		}

		// fix up checked out worktree
		if _, err := repo.executor("restore", "--staged", ".").executeString(); err != nil {
			t.Fatal(err)
		}
		if _, err := repo.executor("checkout", "--", ".").executeString(); err != nil {
			t.Fatal(err)
		}

		mergeTreeID, err := repo.GetMergeTree(commitAID, commitBID)
		assert.Nil(t, err)
		if !combinedTreeID.Equal(mergeTreeID) {
			mergeTreeContents, err := repo.GetAllFilesInTree(mergeTreeID)
			if err != nil {
				t.Fatalf("unexpected error when debugging non-matched merge trees: %s", err.Error())
			}
			t.Log("merge tree contents:", mergeTreeContents)
			t.Error("merge trees don't match")
		}
	})

	t.Run("merge conflict", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, false)

		// We meed to change the directory for this test because we `checkout`
		// for older Git versions, modifying the worktree. This chdir ensures
		// that the temporary directory is used as the worktree.
		pwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(pwd) //nolint:errcheck

		emptyBlobID, err := repo.WriteBlob(nil)
		if err != nil {
			t.Fatal(err)
		}

		treeBuilder := NewTreeBuilder(repo)
		emptyTreeID, err := treeBuilder.WriteTreeFromEntries(nil)
		if err != nil {
			t.Fatal(err)
		}

		blobAID, err := repo.WriteBlob([]byte("a"))
		if err != nil {
			t.Fatal(err)
		}
		blobBID, err := repo.WriteBlob([]byte("b"))
		if err != nil {
			t.Fatal(err)
		}

		treeAID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("a", blobAID)})
		if err != nil {
			t.Fatal(err)
		}
		treeBID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("a", blobBID),
			NewEntryBlob("b", emptyBlobID),
		})
		if err != nil {
			t.Fatal(err)
		}

		mainRef := "refs/heads/main"
		featureRef := "refs/heads/feature"

		// Add commits to the main branch
		baseCommitID, err := repo.Commit(emptyTreeID, mainRef, "Initial commit", false)
		if err != nil {
			t.Fatal(err)
		}
		commitAID, err := repo.Commit(treeAID, mainRef, "Commit A", false)
		if err != nil {
			t.Fatal(err)
		}

		// Add commits to the feature branch
		if err := repo.SetReference(featureRef, baseCommitID); err != nil {
			t.Fatal(err)
		}
		commitBID, err := repo.Commit(treeBID, featureRef, "Commit B", false)
		if err != nil {
			t.Fatal(err)
		}

		// fix up checked out worktree
		if _, err := repo.executor("restore", "--staged", ".").executeString(); err != nil {
			t.Fatal(err)
		}
		if _, err := repo.executor("checkout", "--", ".").executeString(); err != nil {
			t.Fatal(err)
		}

		_, err = repo.GetMergeTree(commitAID, commitBID)
		assert.NotNil(t, err)
	})

	t.Run("fast forward merge", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, false)

		// We meed to change the directory for this test because we `checkout`
		// for older Git versions, modifying the worktree. This chdir ensures
		// that the temporary directory is used as the worktree.
		pwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(pwd) //nolint:errcheck

		emptyBlobID, err := repo.WriteBlob(nil)
		if err != nil {
			t.Fatal(err)
		}

		treeBuilder := NewTreeBuilder(repo)
		treeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("empty", emptyBlobID)})
		if err != nil {
			t.Fatal(err)
		}

		commitID, err := repo.Commit(treeID, "refs/heads/main", "Initial commit\n", false)
		if err != nil {
			t.Fatal(err)
		}

		mergeTreeID, err := repo.GetMergeTree(ZeroHash, commitID)
		assert.Nil(t, err)
		assert.Equal(t, treeID, mergeTreeID)
	})
}

func TestCreateSubtreeFromUpstreamRepository(t *testing.T) {
	t.Run("subtree into HEAD", func(t *testing.T) {
		tmpDir1 := t.TempDir()
		downstreamRepository := CreateTestGitRepository(t, tmpDir1, false)

		blobAID, err := downstreamRepository.WriteBlob([]byte("a"))
		require.Nil(t, err)

		blobBID, err := downstreamRepository.WriteBlob([]byte("b"))
		require.Nil(t, err)

		downstreamTreeBuilder := NewTreeBuilder(downstreamRepository)

		// The downstream tree (if set as exists in test below) is:
		// oof/a -> blobA
		// b     -> blobB
		downstreamTreeEntries := []TreeEntry{
			NewEntryBlob("oof/a", blobAID),
			NewEntryBlob("b", blobBID),
		}
		downstreamTreeID, err := downstreamTreeBuilder.WriteTreeFromEntries(downstreamTreeEntries)
		require.Nil(t, err)

		downstreamCommitID, err := downstreamRepository.Commit(downstreamTreeID, "refs/heads/main", "Initial commit\n", false)
		require.Nil(t, err)

		err = downstreamRepository.SetSymbolicReference("HEAD", "refs/heads/main")
		require.Nil(t, err)

		downstreamRepository.RestoreWorktree(t)

		tmpDir2 := t.TempDir()
		upstreamRepository := CreateTestGitRepository(t, tmpDir2, true)

		_, err = upstreamRepository.WriteBlob([]byte("a"))
		require.Nil(t, err)

		_, err = upstreamRepository.WriteBlob([]byte("b"))
		require.Nil(t, err)

		upstreamTreeBuilder := NewTreeBuilder(upstreamRepository)

		// The upstream tree is:
		// a                -> blobA
		// foo/a            -> blobA
		// foo/b            -> blobB
		// foobar/foo/bar/b -> blobB

		upstreamTreeID, err := upstreamTreeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("a", blobAID),
			NewEntryBlob("foo/a", blobAID),
			NewEntryBlob("foo/b", blobBID),
			NewEntryBlob("foobar/foo/bar/b", blobBID),
		})
		require.Nil(t, err)

		upstreamRef := "refs/heads/main"
		upstreamCommitID, err := upstreamRepository.Commit(upstreamTreeID, upstreamRef, "Initial commit\n", false)
		require.Nil(t, err)

		downstreamCommitIDNew, err := downstreamRepository.CreateSubtreeFromUpstreamRepository(upstreamRepository, upstreamCommitID, "", "refs/heads/main", "upstream")
		assert.Nil(t, err)
		assert.NotEqual(t, downstreamCommitID, downstreamCommitIDNew)

		statuses, err := downstreamRepository.Status()
		require.Nil(t, err)
		assert.Empty(t, statuses)
	})

	t.Run("various other subtree scenarios", func(t *testing.T) {
		tmpDir1 := t.TempDir()
		downstreamRepository := CreateTestGitRepository(t, tmpDir1, false)

		blobAID, err := downstreamRepository.WriteBlob([]byte("a"))
		require.Nil(t, err)

		blobBID, err := downstreamRepository.WriteBlob([]byte("b"))
		require.Nil(t, err)

		downstreamTreeBuilder := NewTreeBuilder(downstreamRepository)

		// The downstream tree (if set as exists in test below) is:
		// oof/a -> blobA
		// b     -> blobB
		downstreamTreeEntries := []TreeEntry{
			NewEntryBlob("oof/a", blobAID),
			NewEntryBlob("b", blobBID),
		}
		downstreamTreeID, err := downstreamTreeBuilder.WriteTreeFromEntries(downstreamTreeEntries)
		require.Nil(t, err)

		tmpDir2 := t.TempDir()
		upstreamRepository := CreateTestGitRepository(t, tmpDir2, true)

		_, err = upstreamRepository.WriteBlob([]byte("a"))
		require.Nil(t, err)

		_, err = upstreamRepository.WriteBlob([]byte("b"))
		require.Nil(t, err)

		upstreamTreeBuilder := NewTreeBuilder(upstreamRepository)

		// The upstream tree is:
		// a                -> blobA
		// foo/a            -> blobA
		// foo/b            -> blobB
		// foobar/foo/bar/b -> blobB

		upstreamRootTreeID, err := upstreamTreeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("a", blobAID),
			NewEntryBlob("foo/a", blobAID),
			NewEntryBlob("foo/b", blobBID),
			NewEntryBlob("foobar/foo/bar/b", blobBID),
		})
		require.Nil(t, err)

		upstreamRef := "refs/heads/main"
		upstreamCommitID, err := upstreamRepository.Commit(upstreamRootTreeID, upstreamRef, "Initial commit\n", false)
		require.Nil(t, err)

		tests := map[string]struct {
			upstreamPath     string
			localPath        string
			refExists        bool // refExists -> we must check for other files but no prior propagation has happened
			priorPropagation bool // priorPropagation -> localPath is already populated, mutually exclusive with refExists
			err              error
		}{
			"directory in root, ref does not exist": {
				localPath:        "upstream",
				refExists:        false,
				priorPropagation: false,
			},
			"directory in root, trailing slash, ref does not exist": {
				localPath:        "upstream/",
				refExists:        false,
				priorPropagation: false,
			},
			"directory in root, ref exists": {
				localPath:        "upstream",
				refExists:        true,
				priorPropagation: false,
			},
			"directory in root, trailing slash, ref exists": {
				localPath:        "upstream/",
				refExists:        true,
				priorPropagation: false,
			},
			"directory in root, prior propagation exists": {
				localPath:        "upstream",
				refExists:        false,
				priorPropagation: true,
			},
			"directory in root, trailing slash, prior propagation exists": {
				localPath:        "upstream/",
				refExists:        false,
				priorPropagation: true,
			},
			"directory in subdirectory, ref does not exist": {
				localPath:        "foo/upstream",
				refExists:        false,
				priorPropagation: false,
			},
			"directory in subdirectory, trailing slash, ref does not exist": {
				localPath:        "foo/upstream/",
				refExists:        false,
				priorPropagation: false,
			},
			"directory in subdirectory, ref exists": {
				localPath:        "foo/upstream",
				refExists:        true,
				priorPropagation: false,
			},
			"directory in subdirectory, trailing slash, ref exists": {
				localPath:        "foo/upstream/",
				refExists:        true,
				priorPropagation: false,
			},
			"directory in subdirectory, prior propagation exists": {
				localPath:        "foo/upstream",
				refExists:        false,
				priorPropagation: true,
			},
			"directory in subdirectory, trailing slash, prior propagation exists": {
				localPath:        "foo/upstream/",
				refExists:        false,
				priorPropagation: true,
			},
			"with upstream path, directory in root, ref does not exist": {
				upstreamPath:     "foo",
				localPath:        "upstream",
				refExists:        false,
				priorPropagation: false,
			},
			"with upstream path, directory in root, trailing slash, ref does not exist": {
				upstreamPath:     "foo/",
				localPath:        "upstream/",
				refExists:        false,
				priorPropagation: false,
			},
			"with upstream path, directory in root, ref exists": {
				upstreamPath:     "foo",
				localPath:        "upstream",
				refExists:        true,
				priorPropagation: false,
			},
			"with upstream path, directory in root, trailing slash, ref exists": {
				upstreamPath:     "foo/",
				localPath:        "upstream/",
				refExists:        true,
				priorPropagation: false,
			},
			"with upstream path, directory in root, prior propagation exists": {
				upstreamPath:     "foo",
				localPath:        "upstream",
				refExists:        false,
				priorPropagation: true,
			},
			"with upstream path, directory in root, trailing slash, prior propagation exists": {
				upstreamPath:     "foo/",
				localPath:        "upstream/",
				refExists:        false,
				priorPropagation: true,
			},
			"with upstream path, directory in subdirectory, ref does not exist": {
				upstreamPath:     "foo",
				localPath:        "foo/upstream",
				refExists:        false,
				priorPropagation: false,
			},
			"with upstream path, directory in subdirectory, trailing slash, ref does not exist": {
				upstreamPath:     "foo/",
				localPath:        "foo/upstream/",
				refExists:        false,
				priorPropagation: false,
			},
			"with upstream path, directory in subdirectory, ref exists": {
				upstreamPath:     "foo",
				localPath:        "foo/upstream",
				refExists:        true,
				priorPropagation: false,
			},
			"with upstream path, directory in subdirectory, trailing slash, ref exists": {
				upstreamPath:     "foo/",
				localPath:        "foo/upstream/",
				refExists:        true,
				priorPropagation: false,
			},
			"with upstream path, directory in subdirectory, prior propagation exists": {
				upstreamPath:     "foo",
				localPath:        "foo/upstream",
				refExists:        false,
				priorPropagation: true,
			},
			"with upstream path, directory in subdirectory, trailing slash, prior propagation exists": {
				upstreamPath:     "foo",
				localPath:        "foo/upstream/",
				refExists:        false,
				priorPropagation: true,
			},
			"with upstream path as subdirectory, directory in root, ref does not exist": {
				upstreamPath:     "foobar/foo",
				localPath:        "upstream",
				refExists:        false,
				priorPropagation: false,
			},
			"with upstream path as subdirectory, directory in root, trailing slash, ref does not exist": {
				upstreamPath:     "foobar/foo/",
				localPath:        "upstream/",
				refExists:        false,
				priorPropagation: false,
			},
			"with upstream path as subdirectory, directory in root, ref exists": {
				upstreamPath:     "foobar/foo",
				localPath:        "upstream",
				refExists:        true,
				priorPropagation: false,
			},
			"with upstream path as subdirectory, directory in root, trailing slash, ref exists": {
				upstreamPath:     "foobar/foo/",
				localPath:        "upstream/",
				refExists:        true,
				priorPropagation: false,
			},
			"with upstream path as subdirectory, directory in root, prior propagation exists": {
				upstreamPath:     "foobar/foo",
				localPath:        "upstream",
				refExists:        false,
				priorPropagation: true,
			},
			"with upstream path as subdirectory, directory in root, trailing slash, prior propagation exists": {
				upstreamPath:     "foobar/foo/",
				localPath:        "upstream/",
				refExists:        false,
				priorPropagation: true,
			},
			"with upstream path as subdirectory, directory in subdirectory, ref does not exist": {
				upstreamPath:     "foobar/foo",
				localPath:        "foo/upstream",
				refExists:        false,
				priorPropagation: false,
			},
			"with upstream path as subdirectory, directory in subdirectory, trailing slash, ref does not exist": {
				upstreamPath:     "foobar/foo/",
				localPath:        "foo/upstream/",
				refExists:        false,
				priorPropagation: false,
			},
			"with upstream path as subdirectory, directory in subdirectory, ref exists": {
				upstreamPath:     "foobar/foo",
				localPath:        "foo/upstream",
				refExists:        true,
				priorPropagation: false,
			},
			"with upstream path as subdirectory, directory in subdirectory, trailing slash, ref exists": {
				upstreamPath:     "foobar/foo/",
				localPath:        "foo/upstream/",
				refExists:        true,
				priorPropagation: false,
			},
			"with upstream path as subdirectory, directory in subdirectory, prior propagation exists": {
				upstreamPath:     "foobar/foo",
				localPath:        "foo/upstream",
				refExists:        false,
				priorPropagation: true,
			},
			"with upstream path as subdirectory, directory in subdirectory, trailing slash, prior propagation exists": {
				upstreamPath:     "foobar/foo/",
				localPath:        "foo/upstream/",
				refExists:        false,
				priorPropagation: true,
			},
			"upstream path does not exist": {
				upstreamPath: "does-not-exist",
				localPath:    "foo/upstream/",
				err:          ErrTreeDoesNotHavePath,
			},
			"empty localPath": {
				err: ErrCannotCreateSubtreeIntoRootTree,
			},
		}

		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				require.False(t, test.refExists && test.priorPropagation, "refExists and priorPropagation can't both be true")

				downstreamRef := testNameToRefName(name)

				if test.refExists {
					_, err := downstreamRepository.Commit(downstreamTreeID, downstreamRef, "Initial commit\n", false)
					require.Nil(t, err)
				} else if test.priorPropagation {
					// We set the upstream path to contain the same tree as the
					// downstreamTree, so:
					// oof/a            -> blobA
					// b                -> blobB
					// <upstream>/oof/a -> blobA
					// <upstream>/b     -> blobB

					entries := []TreeEntry{NewEntryTree(test.localPath, downstreamTreeID)}
					entries = append(entries, downstreamTreeEntries...)

					rootTreeID, err := downstreamTreeBuilder.WriteTreeFromEntries(entries)
					require.Nil(t, err)

					_, err = downstreamRepository.Commit(rootTreeID, downstreamRef, "Initial commit\n", false)
					require.Nil(t, err)
				}

				downstreamCommitID, err := downstreamRepository.CreateSubtreeFromUpstreamRepository(upstreamRepository, upstreamCommitID, test.upstreamPath, downstreamRef, test.localPath)
				if test.err != nil {
					assert.ErrorIs(t, err, test.err)
				} else {
					assert.Nil(t, err)

					rootTreeID, err := downstreamRepository.GetCommitTreeID(downstreamCommitID)
					require.Nil(t, err)

					itemID, err := downstreamRepository.GetPathIDInTree(test.localPath, rootTreeID)
					require.Nil(t, err)

					upstreamTreeID := upstreamRootTreeID
					if test.upstreamPath != "" {
						upstreamTreeID, err = upstreamRepository.GetPathIDInTree(test.upstreamPath, upstreamRootTreeID)
						require.Nil(t, err)
					}
					assert.Equal(t, upstreamTreeID, itemID)

					if test.refExists {
						// check that other items are still present
						itemID, err := downstreamRepository.GetPathIDInTree("oof/a", downstreamTreeID)
						require.Nil(t, err)
						assert.Equal(t, blobAID, itemID)

						itemID, err = downstreamRepository.GetPathIDInTree("b", downstreamTreeID)
						require.Nil(t, err)
						assert.Equal(t, blobBID, itemID)
					}

					// We don't need to similarly check when test.priorPropagation is
					// true because we already checked that those contents don't exist
					// in that localPath when we checked the tree ID patches
					// upstreamTreeID
				}
			})
		}
	})
}

func TestTreeBuilder(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	blobAID, err := repo.WriteBlob([]byte("a"))
	if err != nil {
		t.Fatal(err)
	}

	blobBID, err := repo.WriteBlob([]byte("b"))
	if err != nil {
		t.Fatal(err)
	}

	emptyTreeID := "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

	t.Run("no blobs", func(t *testing.T) {
		treeBuilder := NewTreeBuilder(repo)
		treeID, err := treeBuilder.WriteTreeFromEntries(nil)
		assert.Nil(t, err)
		assert.Equal(t, emptyTreeID, treeID.String())

		treeID, err = treeBuilder.WriteTreeFromEntries(nil)
		assert.Nil(t, err)
		assert.Equal(t, emptyTreeID, treeID.String())
	})

	t.Run("both blobs in the root directory", func(t *testing.T) {
		treeBuilder := NewTreeBuilder(repo)

		input := []TreeEntry{
			NewEntryBlob("a", blobAID),
			NewEntryBlob("b", blobBID),
		}

		rootTreeID, err := treeBuilder.WriteTreeFromEntries(input)
		assert.Nil(t, err)

		files, err := repo.GetAllFilesInTree(rootTreeID)
		if err != nil {
			t.Fatal(err)
		}

		expectedOutput := map[string]Hash{
			"a": blobAID,
			"b": blobBID,
		}
		assert.Equal(t, expectedOutput, files)
	})

	t.Run("both blobs in same subdirectory", func(t *testing.T) {
		treeBuilder := NewTreeBuilder(repo)

		input := []TreeEntry{
			NewEntryBlob("dir/a", blobAID),
			NewEntryBlob("dir/b", blobBID),
		}

		rootTreeID, err := treeBuilder.WriteTreeFromEntries(input)
		assert.Nil(t, err)

		files, err := repo.GetAllFilesInTree(rootTreeID)
		if err != nil {
			t.Fatal(err)
		}

		expectedOutput := map[string]Hash{
			"dir/a": blobAID,
			"dir/b": blobBID,
		}

		assert.Equal(t, expectedOutput, files)
	})

	t.Run("same blobs in the multiple directories", func(t *testing.T) {
		treeBuilder := NewTreeBuilder(repo)

		input := []TreeEntry{
			NewEntryBlob("a", blobAID),
			NewEntryBlob("b", blobBID),
			NewEntryBlob("foo/a", blobAID),
			NewEntryBlob("foo/b", blobBID),
			NewEntryBlob("bar/a", blobAID),
			NewEntryBlob("bar/b", blobBID),
		}

		rootTreeID, err := treeBuilder.WriteTreeFromEntries(input)
		assert.Nil(t, err)

		files, err := repo.GetAllFilesInTree(rootTreeID)
		if err != nil {
			t.Fatal(err)
		}

		expectedOutput := map[string]Hash{
			"a":     blobAID,
			"b":     blobBID,
			"foo/a": blobAID,
			"foo/b": blobBID,
			"bar/a": blobAID,
			"bar/b": blobBID,
		}

		assert.Equal(t, expectedOutput, files)
	})

	t.Run("both blobs in different subdirectories", func(t *testing.T) {
		treeBuilder := NewTreeBuilder(repo)

		input := []TreeEntry{
			NewEntryBlob("foo/a", blobAID),
			NewEntryBlob("bar/b", blobBID),
		}

		rootTreeID, err := treeBuilder.WriteTreeFromEntries(input)
		assert.Nil(t, err)

		files, err := repo.GetAllFilesInTree(rootTreeID)
		if err != nil {
			t.Fatal(err)
		}

		expectedOutput := map[string]Hash{
			"foo/a": blobAID,
			"bar/b": blobBID,
		}

		assert.Equal(t, expectedOutput, files)
	})

	t.Run("blobs in mix of root directory and subdirectories", func(t *testing.T) {
		treeBuilder := NewTreeBuilder(repo)

		input := []TreeEntry{
			NewEntryBlob("a", blobAID),
			NewEntryBlob("foo/bar/foobar/b", blobBID),
		}

		rootTreeID, err := treeBuilder.WriteTreeFromEntries(input)
		assert.Nil(t, err)

		files, err := repo.GetAllFilesInTree(rootTreeID)
		if err != nil {
			t.Fatal(err)
		}

		expectedOutput := map[string]Hash{
			"a":                blobAID,
			"foo/bar/foobar/b": blobBID,
		}

		assert.Equal(t, expectedOutput, files)
	})

	t.Run("build tree from intermediate tree", func(t *testing.T) {
		treeBuilder := NewTreeBuilder(repo)

		intermediateTreeInput := []TreeEntry{
			NewEntryBlob("a", blobAID),
		}

		intermediateTreeID, err := treeBuilder.WriteTreeFromEntries(intermediateTreeInput)
		assert.Nil(t, err)

		rootTreeInput := []TreeEntry{
			NewEntryTree("intermediate", intermediateTreeID),
			NewEntryBlob("b", blobBID),
		}

		rootTreeID, err := treeBuilder.WriteTreeFromEntries(rootTreeInput)
		assert.Nil(t, err)

		expectedFiles := map[string]Hash{
			"intermediate/a": blobAID,
			"b":              blobBID,
		}

		files, err := repo.GetAllFilesInTree(rootTreeID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, expectedFiles, files)
	})

	t.Run("build tree from nested intermediate tree", func(t *testing.T) {
		treeBuilder := NewTreeBuilder(repo)

		intermediateTreeInput := []TreeEntry{
			NewEntryBlob("a", blobAID),
		}

		intermediateTreeID, err := treeBuilder.WriteTreeFromEntries(intermediateTreeInput)
		assert.Nil(t, err)

		rootTreeInput := []TreeEntry{
			NewEntryTree("foo/intermediate", intermediateTreeID),
			NewEntryBlob("b", blobBID),
		}

		rootTreeID, err := treeBuilder.WriteTreeFromEntries(rootTreeInput)
		assert.Nil(t, err)

		expectedFiles := map[string]Hash{
			"foo/intermediate/a": blobAID,
			"b":                  blobBID,
		}

		files, err := repo.GetAllFilesInTree(rootTreeID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, expectedFiles, files)
	})

	t.Run("build tree from nested multi-level intermediate tree", func(t *testing.T) {
		treeBuilder := NewTreeBuilder(repo)

		intermediateTreeInput := []TreeEntry{
			NewEntryBlob("intermediate/a", blobAID),
		}

		intermediateTreeID, err := treeBuilder.WriteTreeFromEntries(intermediateTreeInput)
		assert.Nil(t, err)

		rootTreeInput := []TreeEntry{
			NewEntryTree("foo/intermediate", intermediateTreeID),
			NewEntryBlob("b", blobBID),
		}

		rootTreeID, err := treeBuilder.WriteTreeFromEntries(rootTreeInput)
		assert.Nil(t, err)

		expectedFiles := map[string]Hash{
			"foo/intermediate/intermediate/a": blobAID,
			"b":                               blobBID,
		}

		files, err := repo.GetAllFilesInTree(rootTreeID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, expectedFiles, files)
	})
}

func TestEnsureIsTree(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, true)

	blobID, err := repo.WriteBlob([]byte("foo"))
	if err != nil {
		t.Fatal(err)
	}

	treeBuilder := NewTreeBuilder(repo)
	treeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{NewEntryBlob("foo", blobID)})
	if err != nil {
		t.Fatal(err)
	}

	err = repo.ensureIsTree(treeID)
	assert.Nil(t, err)

	err = repo.ensureIsTree(blobID)
	assert.NotNil(t, err)
}

func TestEnsureIsTreeError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, true)

	// Try with a blob instead of tree
	blobID, err := repo.WriteBlob([]byte("test"))
	require.Nil(t, err)

	err = repo.ensureIsTree(blobID)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "is not a tree object")
}

func TestGetPathIDInTreeError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	treeBuilder := NewTreeBuilder(repo)
	blobID, err := repo.WriteBlob([]byte("test"))
	require.Nil(t, err)

	treeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
		NewEntryBlob("file.txt", blobID),
	})
	require.Nil(t, err)

	// Try to get non-existent path
	_, err = repo.GetPathIDInTree("nonexistent.txt", treeID)
	assert.NotNil(t, err)
}

func TestGetTreeItemsError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Try with a blob instead of tree
	blobID, err := repo.WriteBlob([]byte("test"))
	require.Nil(t, err)

	_, err = repo.GetTreeItems(blobID)
	assert.NotNil(t, err)
}

func TestGetAllFilesInTreeError(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	// Try with a blob instead of tree
	blobID, err := repo.WriteBlob([]byte("test"))
	require.Nil(t, err)

	_, err = repo.GetAllFilesInTree(blobID)
	assert.NotNil(t, err)
}

func TestNewEntryBlobWithPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	blobID, err := repo.WriteBlob([]byte("test"))
	require.Nil(t, err)

	// Test with executable permissions
	entry := NewEntryBlobWithPermissions("script.sh", blobID, 0o755)
	assert.Equal(t, "script.sh", entry.getName())
	assert.Equal(t, blobID, entry.getID())

	// Test with regular file permissions
	entry2 := NewEntryBlobWithPermissions("file.txt", blobID, 0o644)
	assert.Equal(t, "file.txt", entry2.getName())
	assert.Equal(t, blobID, entry2.getID())
}

func TestNewEntryTree(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	treeBuilder := NewTreeBuilder(repo)
	blobID, err := repo.WriteBlob([]byte("test"))
	require.Nil(t, err)

	treeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
		NewEntryBlob("file.txt", blobID),
	})
	require.Nil(t, err)

	entry := NewEntryTree("subdir", treeID)
	assert.Equal(t, "subdir", entry.getName())
	assert.Equal(t, treeID, entry.getID())
}

func TestEmptyTreeEdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	t.Run("get empty tree multiple times", func(t *testing.T) {
		emptyTree1, err := repo.EmptyTree()
		assert.Nil(t, err)

		emptyTree2, err := repo.EmptyTree()
		assert.Nil(t, err)

		// Should return the same hash
		assert.Equal(t, emptyTree1, emptyTree2)
	})

	t.Run("empty tree is valid tree object", func(t *testing.T) {
		emptyTree, err := repo.EmptyTree()
		assert.Nil(t, err)

		objType, err := repo.GetObjectType(emptyTree)
		assert.Nil(t, err)
		assert.Equal(t, TreeObjectType, objType)
	})
}

func TestGetPathIDInTreeEdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	treeBuilder := NewTreeBuilder(repo)
	blobID, err := repo.WriteBlob([]byte("test content"))
	require.Nil(t, err)

	t.Run("nested path in tree", func(t *testing.T) {
		// Create nested tree structure
		subTreeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("file.txt", blobID),
		})
		require.Nil(t, err)

		treeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryTree("subdir", subTreeID),
		})
		require.Nil(t, err)

		// Get nested file
		fileID, err := repo.GetPathIDInTree("subdir/file.txt", treeID)
		assert.Nil(t, err)
		assert.Equal(t, blobID, fileID)
	})

	t.Run("root level file", func(t *testing.T) {
		treeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("root.txt", blobID),
		})
		require.Nil(t, err)

		fileID, err := repo.GetPathIDInTree("root.txt", treeID)
		assert.Nil(t, err)
		assert.Equal(t, blobID, fileID)
	})
}

func TestGetTreeItemsEdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	treeBuilder := NewTreeBuilder(repo)

	t.Run("tree with multiple items", func(t *testing.T) {
		blob1, err := repo.WriteBlob([]byte("content1"))
		require.Nil(t, err)
		blob2, err := repo.WriteBlob([]byte("content2"))
		require.Nil(t, err)
		blob3, err := repo.WriteBlob([]byte("content3"))
		require.Nil(t, err)

		treeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("file1.txt", blob1),
			NewEntryBlob("file2.txt", blob2),
			NewEntryBlob("file3.txt", blob3),
		})
		require.Nil(t, err)

		items, err := repo.GetTreeItems(treeID)
		assert.Nil(t, err)
		assert.Len(t, items, 3)
		assert.Equal(t, blob1, items["file1.txt"])
		assert.Equal(t, blob2, items["file2.txt"])
		assert.Equal(t, blob3, items["file3.txt"])
	})

	t.Run("tree with subdirectories", func(t *testing.T) {
		blobID, err := repo.WriteBlob([]byte("test"))
		require.Nil(t, err)

		subTreeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("nested.txt", blobID),
		})
		require.Nil(t, err)

		treeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryTree("subdir", subTreeID),
			NewEntryBlob("root.txt", blobID),
		})
		require.Nil(t, err)

		items, err := repo.GetTreeItems(treeID)
		assert.Nil(t, err)
		assert.Len(t, items, 2)
		assert.Equal(t, subTreeID, items["subdir"])
		assert.Equal(t, blobID, items["root.txt"])
	})
}

func TestGetAllFilesInTreeEdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	treeBuilder := NewTreeBuilder(repo)

	t.Run("deeply nested files", func(t *testing.T) {
		blobID, err := repo.WriteBlob([]byte("deep content"))
		require.Nil(t, err)

		// Create nested structure: dir1/dir2/dir3/file.txt
		level3Tree, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("file.txt", blobID),
		})
		require.Nil(t, err)

		level2Tree, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryTree("dir3", level3Tree),
		})
		require.Nil(t, err)

		level1Tree, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryTree("dir2", level2Tree),
		})
		require.Nil(t, err)

		rootTree, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryTree("dir1", level1Tree),
		})
		require.Nil(t, err)

		files, err := repo.GetAllFilesInTree(rootTree)
		assert.Nil(t, err)
		assert.Len(t, files, 1)
		assert.Equal(t, blobID, files["dir1/dir2/dir3/file.txt"])
	})

	t.Run("mixed files and directories", func(t *testing.T) {
		blob1, err := repo.WriteBlob([]byte("root content"))
		require.Nil(t, err)
		blob2, err := repo.WriteBlob([]byte("sub content"))
		require.Nil(t, err)

		subTree, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("sub.txt", blob2),
		})
		require.Nil(t, err)

		rootTree, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("root.txt", blob1),
			NewEntryTree("subdir", subTree),
		})
		require.Nil(t, err)

		files, err := repo.GetAllFilesInTree(rootTree)
		assert.Nil(t, err)
		assert.Len(t, files, 2)
		assert.Equal(t, blob1, files["root.txt"])
		assert.Equal(t, blob2, files["subdir/sub.txt"])
	})
}

func TestTreeBuilderComprehensive(t *testing.T) {
	tempDir := t.TempDir()
	repo := CreateTestGitRepository(t, tempDir, false)

	treeBuilder := NewTreeBuilder(repo)

	t.Run("build tree with all entry types", func(t *testing.T) {
		// Create blobs
		blob1, err := repo.WriteBlob([]byte("content1"))
		require.Nil(t, err)
		blob2, err := repo.WriteBlob([]byte("content2"))
		require.Nil(t, err)

		// Create subtree
		subTreeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("subfile.txt", blob1),
		})
		require.Nil(t, err)

		// Create main tree with all types
		entries := []TreeEntry{
			NewEntryBlob("file1.txt", blob1),
			NewEntryBlob("file2.txt", blob2),
			NewEntryTree("subdir", subTreeID),
			NewEntryBlobWithPermissions("executable.sh", blob1, 0o755),
		}

		treeID, err := treeBuilder.WriteTreeFromEntries(entries)
		assert.Nil(t, err)
		assert.False(t, treeID.IsZero())

		// Verify tree items
		items, err := repo.GetTreeItems(treeID)
		assert.Nil(t, err)
		assert.Len(t, items, 4)
	})

	t.Run("build deeply nested tree structure", func(t *testing.T) {
		blobID, err := repo.WriteBlob([]byte("deep"))
		require.Nil(t, err)

		// Build 10 levels deep
		currentTree := blobID
		for i := 0; i < 10; i++ {
			if i == 0 {
				currentTree, err = treeBuilder.WriteTreeFromEntries([]TreeEntry{
					NewEntryBlob(fmt.Sprintf("file%d.txt", i), blobID),
				})
			} else {
				currentTree, err = treeBuilder.WriteTreeFromEntries([]TreeEntry{
					NewEntryTree(fmt.Sprintf("level%d", i), currentTree),
				})
			}
			require.Nil(t, err)
		}

		assert.False(t, currentTree.IsZero())
	})

	t.Run("build tree with many files", func(t *testing.T) {
		var entries []TreeEntry
		for i := 0; i < 50; i++ {
			blobID, err := repo.WriteBlob([]byte(fmt.Sprintf("content%d", i)))
			require.Nil(t, err)
			entries = append(entries, NewEntryBlob(fmt.Sprintf("file%d.txt", i), blobID))
		}

		treeID, err := treeBuilder.WriteTreeFromEntries(entries)
		assert.Nil(t, err)

		items, err := repo.GetTreeItems(treeID)
		assert.Nil(t, err)
		assert.Len(t, items, 50)
	})
}

func TestGetMergeTreeComprehensive(t *testing.T) {
	t.Run("merge tree with no conflicts", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo := CreateTestGitRepository(t, tmpDir, false)

		treeBuilder := NewTreeBuilder(repo)
		emptyTree, err := treeBuilder.WriteTreeFromEntries(nil)
		require.Nil(t, err)

		// Create base commit
		baseCommit, err := repo.Commit(emptyTree, "refs/heads/main", "Base\n", false)
		require.Nil(t, err)

		// Create branch A
		blobA, err := repo.WriteBlob([]byte("A"))
		require.Nil(t, err)
		treeA, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("fileA.txt", blobA),
		})
		require.Nil(t, err)
		commitA, err := repo.Commit(treeA, "refs/heads/branch-a", "A\n", false)
		require.Nil(t, err)

		// Create branch B from base
		err = repo.SetReference("refs/heads/branch-b", baseCommit)
		require.Nil(t, err)
		blobB, err := repo.WriteBlob([]byte("B"))
		require.Nil(t, err)
		treeB, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
			NewEntryBlob("fileB.txt", blobB),
		})
		require.Nil(t, err)
		commitB, err := repo.Commit(treeB, "refs/heads/branch-b", "B\n", false)
		require.Nil(t, err)

		// Get merge tree
		mergeTree, err := repo.GetMergeTree(commitA, commitB)
		assert.Nil(t, err)
		assert.False(t, mergeTree.IsZero())

		// Verify merge tree contains both files
		files, err := repo.GetAllFilesInTree(mergeTree)
		assert.Nil(t, err)
		assert.Len(t, files, 2)
	})
}

func TestTreeEntryTypes(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	blobID, err := repo.WriteBlob([]byte("test"))
	require.Nil(t, err)

	treeBuilder := NewTreeBuilder(repo)
	subTreeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
		NewEntryBlob("file.txt", blobID),
	})
	require.Nil(t, err)

	t.Run("blob entry", func(t *testing.T) {
		entry := NewEntryBlob("test.txt", blobID)
		assert.Equal(t, "test.txt", entry.getName())
		assert.Equal(t, blobID, entry.getID())
	})

	t.Run("tree entry", func(t *testing.T) {
		entry := NewEntryTree("subdir", subTreeID)
		assert.Equal(t, "subdir", entry.getName())
		assert.Equal(t, subTreeID, entry.getID())
	})

	t.Run("executable entry with permissions", func(t *testing.T) {
		entry := NewEntryBlobWithPermissions("script.sh", blobID, 0o755)
		assert.Equal(t, "script.sh", entry.getName())
		assert.Equal(t, blobID, entry.getID())
	})
}

func TestTreeBuilderWithNestedStructure(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	treeBuilder := NewTreeBuilder(repo)

	// Create blobs
	blob1, err := repo.WriteBlob([]byte("content1"))
	require.Nil(t, err)

	blob2, err := repo.WriteBlob([]byte("content2"))
	require.Nil(t, err)

	blob3, err := repo.WriteBlob([]byte("content3"))
	require.Nil(t, err)

	// Create nested structure: dir1/dir2/file.txt
	treeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
		NewEntryBlob("root.txt", blob1),
		NewEntryBlob("dir1/file1.txt", blob2),
		NewEntryBlob("dir1/dir2/file2.txt", blob3),
	})
	require.Nil(t, err)
	assert.False(t, treeID.IsZero())

	// Verify all files are in the tree
	files, err := repo.GetAllFilesInTree(treeID)
	assert.Nil(t, err)
	assert.Len(t, files, 3)
	assert.Contains(t, files, "root.txt")
	assert.Contains(t, files, "dir1/file1.txt")
	assert.Contains(t, files, "dir1/dir2/file2.txt")
}

func TestGetAllFilesInTreeWithExecutables(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	treeBuilder := NewTreeBuilder(repo)

	blob1, err := repo.WriteBlob([]byte("#!/bin/bash\necho hello"))
	require.Nil(t, err)

	blob2, err := repo.WriteBlob([]byte("regular file"))
	require.Nil(t, err)

	treeID, err := treeBuilder.WriteTreeFromEntries([]TreeEntry{
		NewEntryBlobWithPermissions("script.sh", blob1, 0o755),
		NewEntryBlob("readme.txt", blob2),
	})
	require.Nil(t, err)

	files, err := repo.GetAllFilesInTree(treeID)
	assert.Nil(t, err)
	assert.Len(t, files, 2)
	assert.Contains(t, files, "script.sh")
	assert.Contains(t, files, "readme.txt")
}

func TestTreeBuilderWithManyFiles(t *testing.T) {
	tmpDir := t.TempDir()
	repo := CreateTestGitRepository(t, tmpDir, false)

	treeBuilder := NewTreeBuilder(repo)

	// Create many files
	entries := make([]TreeEntry, 50)
	for i := 0; i < 50; i++ {
		blobID, err := repo.WriteBlob([]byte(fmt.Sprintf("content%d", i)))
		require.Nil(t, err)

		entries[i] = NewEntryBlob(fmt.Sprintf("file%d.txt", i), blobID)
	}

	treeID, err := treeBuilder.WriteTreeFromEntries(entries)
	assert.Nil(t, err)
	assert.False(t, treeID.IsZero())

	// Verify all files
	files, err := repo.GetAllFilesInTree(treeID)
	assert.Nil(t, err)
	assert.Len(t, files, 50)
}
