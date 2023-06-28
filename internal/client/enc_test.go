package client

import (
	"crypto/rand"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUserSecretSerialization(t *testing.T) {
	tests := []struct {
		name string
		key  []byte
	}{
		{
			name: "fixed 1",
			key:  []byte("12345678901234567890123456789012"),
		},
		{
			name: "fixed 2",
			key:  []byte("123456aaa0123bbb78901ccc567ddd12"),
		},
		{
			name: "random 1",
			key:  nil,
		},
		{
			name: "random 2",
			key:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.key
			if key == nil {
				// generate random key
				key = make([]byte, 32)
				_, err := rand.Read(key)
				require.NoError(t, err, "rand.Read")
			}

			userSecret := &UserSecret{data: key}
			marshaledData1, err := json.Marshal(userSecret)
			require.NoError(t, err, "json.Marshal")

			var unmarshaledSecret UserSecret
			require.NoError(t, json.Unmarshal(marshaledData1, &unmarshaledSecret), "json.Unmarshal")

			marshaledData2, err := json.Marshal(&unmarshaledSecret)
			require.NoError(t, err, "json.Marshal")

			assert.Equal(t, marshaledData1, marshaledData2, "marshaled data should be equal")
			assert.Equal(t, key, unmarshaledSecret.data, "unmarshaled data should be equal")
		})
	}
}

func TestUserSecretEncryption(t *testing.T) {
	data := []byte("some data to be encrypted")
	tests := []struct {
		name string
		key  []byte
	}{
		{
			name: "fixed 1",
			key:  []byte("12345678901234567890123456789012"),
		},
		{
			name: "fixed 2",
			key:  []byte("123456aaa0123bbb78901ccc567ddd12"),
		},
		{
			name: "random 1",
			key:  nil,
		},
		{
			name: "random 2",
			key:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.key
			if key == nil {
				// generate random key
				key = make([]byte, 32)
				_, err := rand.Read(key)
				require.NoError(t, err, "rand.Read")
			}

			nonce := make([]byte, 12)
			_, err := rand.Read(nonce)
			require.NoError(t, err, "rand.Read")

			userSecret := &UserSecret{data: key}
			encryptedData, err := userSecret.Encrypt(data, nonce)
			require.NoError(t, err, "userSecret.Encrypt")

			decryptedData, err := userSecret.Decrypt(encryptedData, nonce)
			require.NoError(t, err, "userSecret.Decrypt")

			assert.Equal(t, data, decryptedData, "decrypted data should be equal")

			// try to decrypt with different nonce
			nonce = make([]byte, 12)
			_, err = rand.Read(nonce)
			require.NoError(t, err, "rand.Read")

			decryptedData, err = userSecret.Decrypt(encryptedData, nonce)
			assert.Error(t, err, "data should not be decrypted with different nonce")
		})
	}
}

func TestUserIDConsistent(t *testing.T) {
	// generate random key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err, "rand.Read")

	userSecret := &UserSecret{data: key}

	userID1 := userSecret.UserId()
	userID2 := userSecret.UserId()

	assert.Equal(t, userID1, userID2, "user id should be consistent")
}
