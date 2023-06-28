package client

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gihtub.com/lsmoura/hishtory_cc/internal/model"
	"github.com/rs/xid"
	"io"
	"time"
)

type UserSecret struct {
	data []byte
}

func NewUserSecret() (*UserSecret, error) {
	secret := make([]byte, 32)

	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("rand.Read: %w", err)
	}
	return &UserSecret{secret}, nil
}

func (u *UserSecret) IsEmpty() bool {
	if u == nil {
		return true
	}
	return len(u.data) == 0
}

func (u *UserSecret) UserId() string {
	secret := u.data

	h := hmac.New(sha256.New, secret)
	h.Write([]byte("user_id"))
	secretKey := h.Sum(nil)

	return base64.StdEncoding.EncodeToString(secretKey)
}

func (u *UserSecret) EncryptionKey() []byte {
	secret := u.data

	h := hmac.New(sha256.New, secret)
	h.Write([]byte("encryption_key"))

	return sha256hmac(secret, []byte("encryption_key"))
}

func (u *UserSecret) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == "" {
		return nil
	}

	resp, err := base64.StdEncoding.DecodeString(string(data[1 : len(data)-1]))
	if err != nil {
		return fmt.Errorf("failed to decode user secret: %w", err)
	}

	u.data = resp

	return nil
}

func (u *UserSecret) MarshalJSON() ([]byte, error) {
	if u == nil || len(u.data) == 0 {
		return []byte("null"), nil
	}

	encodedString := base64.StdEncoding.EncodeToString(u.data)

	resp := append([]byte(`"`), []byte(encodedString)...)
	resp = append(resp, []byte(`"`)...)

	return resp, nil
}

func (u *UserSecret) Encrypt(data, nonce []byte) ([]byte, error) {
	return EncryptData(u.EncryptionKey(), data, nonce)
}

func (u *UserSecret) Decrypt(data, nonce []byte) ([]byte, error) {
	return DecryptData(u.EncryptionKey(), data, []byte{}, nonce)
}

func sha256hmac(key, additionalData []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(additionalData)
	return h.Sum(nil)
}

func EncryptionKeyFromSecret(userSecret []byte) []byte {
	return sha256hmac(userSecret, []byte("encryption_key"))
}

func makeAEAD(userSecret []byte) (cipher.AEAD, error) {
	key := EncryptionKeyFromSecret(userSecret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return aead, nil
}

func UserIdFromSecret(secret *UserSecret) string {
	h := hmac.New(sha256.New, secret.data)
	h.Write([]byte("user_id"))
	secretKey := h.Sum(nil)

	return base64.StdEncoding.EncodeToString(secretKey)
}

func Nonce() ([]byte, error) {
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to read a nonce: %w", err)
	}

	return nonce, nil
}

func EncryptData(userSecret, data, nonce []byte) ([]byte, error) {
	aead, err := makeAEAD(userSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to make AEAD: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, data, []byte{})

	return ciphertext, nil
}

func DecryptData(userSecret, data, additionalData, nonce []byte) ([]byte, error) {
	aead, err := makeAEAD(userSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to make AEAD: %w", err)
	}
	plaintext, err := aead.Open(nil, nonce, data, additionalData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}
	return plaintext, nil
}

func Encrypt(encryptionKey []byte, userID string, entry *model.HistoryEntry) (*model.EncHistoryEntry, error) {
	serializedData, err := json.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	nonce, err := Nonce()
	if err != nil {
		return nil, fmt.Errorf("Nonce: %w", err)
	}
	bytes, err := EncryptData(encryptionKey, serializedData, nonce)

	return &model.EncHistoryEntry{
		EncryptedData: bytes,
		Nonce:         nonce,
		DeviceID:      entry.DeviceID,
		UserID:        userID,
		Date:          time.Now(),
		EncryptedID:   xid.New().String(),
		ReadCount:     0,
	}, nil
}

func Decrypt(encryptionKey []byte, encEntry *model.EncHistoryEntry) (*model.HistoryEntry, error) {
	serializedData, err := DecryptData(encryptionKey, encEntry.EncryptedData, []byte{}, encEntry.Nonce)
	if err != nil {
		return nil, fmt.Errorf("DecryptData: %w", err)
	}

	var entry model.HistoryEntry
	if err := json.Unmarshal(serializedData, &entry); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return &entry, nil
}
