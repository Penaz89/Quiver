package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/argon2"
)

type Secret struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Username string `json:"username"`
	Password string `json:"password"`
	Notes    string `json:"notes"`
}

type EncryptedVault struct {
	Salt []byte `json:"salt"`
	Data []byte `json:"data"` // AES-GCM encrypted JSON of []Secret
}

const (
	vaultFile  = "vault.json"
	saltLen    = 16
	keyLen     = 32
	timeCost   = 1
	memoryCost = 64 * 1024
	threads    = 4
)

func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, timeCost, uint32(memoryCost), uint8(threads), uint32(keyLen))
}

func VaultExists(dataDir string) bool {
	_, err := os.Stat(filepath.Join(dataDir, vaultFile))
	return err == nil
}

func DeleteVault(dataDir string) error {
	return os.Remove(filepath.Join(dataDir, vaultFile))
}

func OpenVault(dataDir, masterPwd string) ([]Secret, error) {
	path := filepath.Join(dataDir, vaultFile)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Secret{}, nil
		}
		return nil, err
	}

	var ev EncryptedVault
	if err := json.Unmarshal(b, &ev); err != nil {
		return nil, err
	}

	key := DeriveKey(masterPwd, ev.Salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ev.Data) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	nonce, ciphertext := ev.Data[:gcm.NonceSize()], ev.Data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("invalid master password or corrupted data")
	}

	var secrets []Secret
	if err := json.Unmarshal(plaintext, &secrets); err != nil {
		return nil, err
	}

	return secrets, nil
}

func SaveVault(dataDir, masterPwd string, secrets []Secret) error {
	path := filepath.Join(dataDir, vaultFile)

	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}

	key := DeriveKey(masterPwd, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	plaintext, err := json.Marshal(secrets)
	if err != nil {
		return err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	ev := EncryptedVault{
		Salt: salt,
		Data: ciphertext,
	}

	b, err := json.MarshalIndent(ev, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0600)
}
