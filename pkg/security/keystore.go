// Package security provides secure key storage for GopherStrike
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	
	"golang.org/x/crypto/pbkdf2"
)

// SecureKeyStore provides encrypted storage for API keys and sensitive data
type SecureKeyStore struct {
	filePath   string
	masterKey  []byte
	data       map[string]string
	mutex      sync.RWMutex
	gcm        cipher.AEAD
}

// EncryptedData represents the structure of encrypted data stored on disk
type EncryptedData struct {
	Salt           string            `json:"salt"`
	Nonce          string            `json:"nonce"`
	EncryptedKeys  string            `json:"encrypted_keys"`
	KeyDerivationIterations int      `json:"key_derivation_iterations"`
	Version        int               `json:"version"`
}

// NewSecureKeyStore creates a new secure key store
func NewSecureKeyStore(filePath string, password string) (*SecureKeyStore, error) {
	store := &SecureKeyStore{
		filePath: filePath,
		data:     make(map[string]string),
	}
	
	// Derive encryption key from password
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	
	// Use PBKDF2 with SHA-256 for key derivation
	iterations := 100000 // NIST recommended minimum
	store.masterKey = pbkdf2.Key([]byte(password), salt, iterations, 32, sha256.New)
	
	// Create AES-GCM cipher
	block, err := aes.NewCipher(store.masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	
	store.gcm, err = cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	
	// Try to load existing data
	if err := store.loadFromFile(password); err != nil {
		// If file doesn't exist, that's okay - we'll create it on first save
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load existing keystore: %w", err)
		}
	}
	
	return store, nil
}

// Set stores a key-value pair securely
func (ks *SecureKeyStore) Set(key, value string) error {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()
	
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	
	ks.data[key] = value
	return ks.saveToFile()
}

// Get retrieves a value by key
func (ks *SecureKeyStore) Get(key string) (string, error) {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()
	
	value, exists := ks.data[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}
	
	return value, nil
}

// Delete removes a key-value pair
func (ks *SecureKeyStore) Delete(key string) error {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()
	
	delete(ks.data, key)
	return ks.saveToFile()
}

// List returns all keys (but not values)
func (ks *SecureKeyStore) List() []string {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()
	
	keys := make([]string, 0, len(ks.data))
	for key := range ks.data {
		keys = append(keys, key)
	}
	
	return keys
}

// Exists checks if a key exists
func (ks *SecureKeyStore) Exists(key string) bool {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()
	
	_, exists := ks.data[key]
	return exists
}

// saveToFile encrypts and saves the data to file
func (ks *SecureKeyStore) saveToFile() error {
	// Ensure directory exists with secure permissions
	dir := filepath.Dir(ks.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Marshal the data
	jsonData, err := json.Marshal(ks.data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	
	// Generate a random nonce for this encryption
	nonce := make([]byte, ks.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}
	
	// Encrypt the data
	ciphertext := ks.gcm.Seal(nil, nonce, jsonData, nil)
	
	// Generate new salt for this save
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}
	
	// Create the encrypted data structure
	encData := EncryptedData{
		Salt:                    base64.StdEncoding.EncodeToString(salt),
		Nonce:                   base64.StdEncoding.EncodeToString(nonce),
		EncryptedKeys:           base64.StdEncoding.EncodeToString(ciphertext),
		KeyDerivationIterations: 100000,
		Version:                 1,
	}
	
	// Marshal the encrypted structure
	encryptedJSON, err := json.MarshalIndent(encData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted data: %w", err)
	}
	
	// Write to file with secure permissions (owner read/write only)
	if err := os.WriteFile(ks.filePath, encryptedJSON, 0600); err != nil {
		return fmt.Errorf("failed to write keystore file: %w", err)
	}
	
	return nil
}

// loadFromFile decrypts and loads data from file
func (ks *SecureKeyStore) loadFromFile(password string) error {
	// Read the encrypted file
	encryptedData, err := os.ReadFile(ks.filePath)
	if err != nil {
		return err
	}
	
	// Parse the encrypted data structure
	var encData EncryptedData
	if err := json.Unmarshal(encryptedData, &encData); err != nil {
		return fmt.Errorf("failed to parse encrypted data: %w", err)
	}
	
	// Decode base64 components
	salt, err := base64.StdEncoding.DecodeString(encData.Salt)
	if err != nil {
		return fmt.Errorf("failed to decode salt: %w", err)
	}
	
	nonce, err := base64.StdEncoding.DecodeString(encData.Nonce)
	if err != nil {
		return fmt.Errorf("failed to decode nonce: %w", err)
	}
	
	ciphertext, err := base64.StdEncoding.DecodeString(encData.EncryptedKeys)
	if err != nil {
		return fmt.Errorf("failed to decode ciphertext: %w", err)
	}
	
	// Derive the decryption key
	derivedKey := pbkdf2.Key([]byte(password), salt, encData.KeyDerivationIterations, 32, sha256.New)
	
	// Create cipher with derived key
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return fmt.Errorf("failed to create cipher for decryption: %w", err)
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM for decryption: %w", err)
	}
	
	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt data (wrong password?): %w", err)
	}
	
	// Parse the decrypted JSON
	if err := json.Unmarshal(plaintext, &ks.data); err != nil {
		return fmt.Errorf("failed to parse decrypted data: %w", err)
	}
	
	// Update master key for future operations
	ks.masterKey = derivedKey
	ks.gcm = gcm
	
	return nil
}

// ChangePassword changes the encryption password
func (ks *SecureKeyStore) ChangePassword(oldPassword, newPassword string) error {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()
	
	// Verify old password by trying to decrypt
	tempStore := &SecureKeyStore{filePath: ks.filePath}
	if err := tempStore.loadFromFile(oldPassword); err != nil {
		return fmt.Errorf("incorrect old password: %w", err)
	}
	
	// Generate new key with new password
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}
	
	newKey := pbkdf2.Key([]byte(newPassword), salt, 100000, 32, sha256.New)
	
	// Create new cipher
	block, err := aes.NewCipher(newKey)
	if err != nil {
		return fmt.Errorf("failed to create new cipher: %w", err)
	}
	
	newGCM, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create new GCM: %w", err)
	}
	
	// Update the store
	ks.masterKey = newKey
	ks.gcm = newGCM
	
	// Save with new encryption
	return ks.saveToFile()
}

// Backup creates an encrypted backup of the keystore
func (ks *SecureKeyStore) Backup(backupPath string) error {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()
	
	// Read current encrypted file
	data, err := os.ReadFile(ks.filePath)
	if err != nil {
		return fmt.Errorf("failed to read keystore: %w", err)
	}
	
	// Write backup with secure permissions
	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}
	
	return nil
}

// Wipe securely deletes the keystore file and clears memory
func (ks *SecureKeyStore) Wipe() error {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()
	
	// Clear in-memory data
	for key := range ks.data {
		delete(ks.data, key)
	}
	
	// Clear master key
	for i := range ks.masterKey {
		ks.masterKey[i] = 0
	}
	
	// Remove file
	if err := os.Remove(ks.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove keystore file: %w", err)
	}
	
	return nil
}

// GetAPIKeysFromEnvironment loads API keys from environment variables
func GetAPIKeysFromEnvironment() map[string]string {
	apiKeys := make(map[string]string)
	
	// Common API key environment variables
	envVars := []string{
		"SHODAN_API_KEY",
		"VIRUSTOTAL_API_KEY",
		"CENSYS_API_ID",
		"CENSYS_API_SECRET",
		"GITHUB_TOKEN",
		"SECURITY_TRAILS_API_KEY",
		"HUNTER_API_KEY",
	}
	
	for _, envVar := range envVars {
		if value := os.Getenv(envVar); value != "" {
			apiKeys[envVar] = value
		}
	}
	
	return apiKeys
}

// ValidateAPIKey performs basic validation on API key format
func ValidateAPIKey(service, key string) error {
	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	
	switch service {
	case "shodan":
		if len(key) != 32 {
			return fmt.Errorf("Shodan API keys should be 32 characters")
		}
	case "virustotal":
		if len(key) != 64 {
			return fmt.Errorf("VirusTotal API keys should be 64 characters")
		}
	case "censys":
		// Censys uses UUID format
		if len(key) != 36 {
			return fmt.Errorf("Censys API keys should be 36 characters (UUID format)")
		}
	}
	
	return nil
}