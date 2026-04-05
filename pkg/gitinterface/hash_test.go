// Copyright The gittuf Authors
// SPDX-License-Identifier: Apache-2.0

package gitinterface

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHash(t *testing.T) {
	tests := map[string]struct {
		hash          string
		expectedError error
	}{
		"correctly encoded SHA-1 hash": {
			hash: "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
		},
		"correctly encoded SHA-256 hash": {
			hash: "61658570165bc04af68cef20d72da49b070dc9d8cd7c8a526c950b658f4d3ccf",
		},
		"correctly encoded SHA-1 zero hash": {
			hash: "0000000000000000000000000000000000000000",
		},
		"correctly encoded SHA-256 zero hash": {
			hash: "0000000000000000000000000000000000000000000000000000000000000000",
		},
		"incorrect length SHA-1 hash": {
			hash:          "e69de29bb2d1d6434b8",
			expectedError: ErrInvalidHashLength,
		},
		"incorrect length SHA-256 hash": {
			hash:          "61658570165bc04af68cef20d72da49b070dc9d8cd7c8a526c950b658f4d3ccfabcdef",
			expectedError: ErrInvalidHashLength,
		},
		"incorrectly encoded SHA-1 hash": {
			hash:          "e69de29bb2d1d6434b8b29ae775ad8c2e48c539g", // last char is 'g'
			expectedError: ErrInvalidHashEncoding,
		},
		"incorrectly encoded SHA-256 hash": {
			hash:          "61658570165bc04af68cef20d72da49b070dc9d8cd7c8a526c950b658f4d3ccg", // last char is 'g'
			expectedError: ErrInvalidHashEncoding,
		},
	}

	for name, test := range tests {
		hash, err := NewHash(test.hash)
		if test.expectedError == nil {
			expectedHash, secErr := hex.DecodeString(test.hash)
			require.Nil(t, secErr)

			assert.Equal(t, Hash(expectedHash), hash)
			assert.Equal(t, test.hash, hash.String())
			assert.Nil(t, err, fmt.Sprintf("unexpected error in test '%s'", name))
		} else {
			assert.ErrorIs(t, err, test.expectedError, fmt.Sprintf("unexpected error in test '%s'", name))
		}
	}
}

func TestHashIsZero(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		expected bool
	}{
		{"SHA-1 zero hash", "0000000000000000000000000000000000000000", true},
		{"SHA-256 zero hash", "0000000000000000000000000000000000000000000000000000000000000000", true},
		{"SHA-1 non-zero hash", "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391", false},
		{"SHA-256 non-zero hash", "61658570165bc04af68cef20d72da49b070dc9d8cd7c8a526c950b658f4d3ccf", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := NewHash(tt.hash)
			require.Nil(t, err)
			result := hash.IsZero()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashEqual(t *testing.T) {
	hash1, err := NewHash("e69de29bb2d1d6434b8b29ae775ad8c2e48c5391")
	require.Nil(t, err)
	hash2, err := NewHash("e69de29bb2d1d6434b8b29ae775ad8c2e48c5391")
	require.Nil(t, err)
	hash3, err := NewHash("61658570165bc04af68cef20d72da49b070dc9d8cd7c8a526c950b658f4d3ccf")
	require.Nil(t, err)

	tests := []struct {
		name     string
		hash1    Hash
		hash2    Hash
		expected bool
	}{
		{"Same SHA-1 hashes", hash1, hash2, true},
		{"Different hashes", hash1, hash3, false},
		{"Hash compared to itself", hash1, hash1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.hash1.Equal(tt.hash2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashString(t *testing.T) {
	tests := []struct {
		name     string
		hashStr  string
		expected string
	}{
		{"SHA-1 hash", "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391", "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"},
		{"SHA-256 hash", "61658570165bc04af68cef20d72da49b070dc9d8cd7c8a526c950b658f4d3ccf", "61658570165bc04af68cef20d72da49b070dc9d8cd7c8a526c950b658f4d3ccf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := NewHash(tt.hashStr)
			require.Nil(t, err)
			result := hash.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestZeroHash(t *testing.T) {
	assert.True(t, ZeroHash.IsZero())
	assert.Equal(t, "0000000000000000000000000000000000000000", ZeroHash.String())
}

func TestHashIsZeroAllCases(t *testing.T) {
	// Test SHA-1 zero hash
	sha1Zero, err := NewHash("0000000000000000000000000000000000000000")
	require.Nil(t, err)
	assert.True(t, sha1Zero.IsZero())

	// Test SHA-256 zero hash
	sha256Zero, err := NewHash("0000000000000000000000000000000000000000000000000000000000000000")
	require.Nil(t, err)
	assert.True(t, sha256Zero.IsZero())

	// Test non-zero SHA-1 hash
	sha1NonZero, err := NewHash("e69de29bb2d1d6434b8b29ae775ad8c2e48c5391")
	require.Nil(t, err)
	assert.False(t, sha1NonZero.IsZero())

	// Test non-zero SHA-256 hash
	sha256NonZero, err := NewHash("61658570165bc04af68cef20d72da49b070dc9d8cd7c8a526c950b658f4d3ccf")
	require.Nil(t, err)
	assert.False(t, sha256NonZero.IsZero())
}

func TestHashEqualAllCases(t *testing.T) {
	hash1, err := NewHash("e69de29bb2d1d6434b8b29ae775ad8c2e48c5391")
	require.Nil(t, err)

	hash2, err := NewHash("e69de29bb2d1d6434b8b29ae775ad8c2e48c5391")
	require.Nil(t, err)

	hash3, err := NewHash("61658570165bc04af68cef20d72da49b070dc9d8cd7c8a526c950b658f4d3ccf")
	require.Nil(t, err)

	// Test equal hashes
	assert.True(t, hash1.Equal(hash2))
	assert.True(t, hash2.Equal(hash1))

	// Test unequal hashes
	assert.False(t, hash1.Equal(hash3))
	assert.False(t, hash3.Equal(hash1))

	// Test hash equal to another with same value
	assert.True(t, hash1.Equal(hash2))
}

func TestNewHashErrorCases(t *testing.T) {
	// Test invalid length
	_, err := NewHash("abc")
	assert.ErrorIs(t, err, ErrInvalidHashLength)

	// Test invalid encoding
	_, err = NewHash("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	assert.ErrorIs(t, err, ErrInvalidHashEncoding)

	// Test empty string
	_, err = NewHash("")
	assert.ErrorIs(t, err, ErrInvalidHashLength)
}

func TestHashOperations(t *testing.T) {
	// Test various hash operations to ensure all code paths are covered
	hashes := []string{
		"e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
		"61658570165bc04af68cef20d72da49b070dc9d8cd7c8a526c950b658f4d3ccf",
		"0000000000000000000000000000000000000000",
		"0000000000000000000000000000000000000000000000000000000000000000",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}

	for _, hashStr := range hashes {
		hash, err := NewHash(hashStr)
		assert.Nil(t, err)

		// Test String()
		result := hash.String()
		assert.Equal(t, hashStr, result)

		// Test IsZero()
		isZero := hash.IsZero()
		if hashStr == "0000000000000000000000000000000000000000" || hashStr == "0000000000000000000000000000000000000000000000000000000000000000" {
			assert.True(t, isZero)
		} else {
			assert.False(t, isZero)
		}

		// Test Equal()
		hash2, err := NewHash(hashStr)
		assert.Nil(t, err)
		assert.True(t, hash.Equal(hash2))
	}
}

func TestNewHashWithInvalidLength(t *testing.T) {
	// Test with string that's too short
	_, err := NewHash("abc")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "wrong length")

	// Test with string that's too long
	_, err = NewHash("0123456789abcdef0123456789abcdef0123456789abcdef")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "wrong length")

	// Test with invalid hex characters
	_, err = NewHash("ghijklmnopqrstuvwxyzghijklmnopqr")
	assert.NotNil(t, err)
}

func TestHashStringRepresentation(t *testing.T) {
	// Test that String() returns lowercase hex
	hash, err := NewHash("ABCDEF0123456789ABCDEF0123456789ABCDEF01")
	require.Nil(t, err)

	str := hash.String()
	assert.Equal(t, "abcdef0123456789abcdef0123456789abcdef01", str)
	assert.Equal(t, 40, len(str))
}

func TestHashComparison(t *testing.T) {
	hash1, _ := NewHash("0000000000000000000000000000000000000001")
	hash2, _ := NewHash("0000000000000000000000000000000000000001")
	hash3, _ := NewHash("0000000000000000000000000000000000000002")

	// Test Equal
	assert.True(t, hash1.Equal(hash2))
	assert.False(t, hash1.Equal(hash3))
	assert.False(t, hash1.Equal(ZeroHash))

	// Test with ZeroHash
	assert.True(t, ZeroHash.IsZero())
	assert.False(t, hash1.IsZero())
	assert.False(t, ZeroHash.Equal(hash1))
}

func TestZeroHashOperations(t *testing.T) {
	// Test ZeroHash is actually zero
	assert.True(t, ZeroHash.IsZero())

	// Test non-zero hash
	hash, _ := NewHash("abc123def456789012345678901234567890abcd")
	assert.False(t, hash.IsZero())

	// Test ZeroHash string representation
	assert.Equal(t, "0000000000000000000000000000000000000000", ZeroHash.String())
}
