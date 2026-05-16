package auth

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const keyringService = "fiken-go"

// KeyringSource reads a token from the OS keyring under
// (service="fiken-go", user=profile). Falls back to
// ~/.config/fiken/credentials/<profile>.json (mode 0600) when the
// keyring is unavailable.
type KeyringSource struct {
	Profile  string
	FilePath string // override; default ~/.config/fiken/credentials/<profile>.json
}

// Token implements Source.
func (k KeyringSource) Token(_ context.Context) (string, error) {
	raw, err := keyring.Get(keyringService, k.Profile)
	if err == nil {
		return tokenFromRaw(raw)
	}
	fp := k.FilePath
	if fp == "" {
		fp = defaultFilePath(k.Profile)
	}
	data, ferr := os.ReadFile(fp) // #nosec G304 -- credential file path is operator-configured
	if ferr != nil {
		if errors.Is(ferr, os.ErrNotExist) {
			return "", ErrNotFound
		}
		return "", ferr
	}
	return tokenFromRaw(string(data))
}

// Save stores cred in the keyring, falling back to file if keyring
// is unavailable. Returns "keyring" or the file path used.
func (k KeyringSource) Save(cred Credential) (string, error) {
	raw, err := json.Marshal(cred)
	if err != nil {
		return "", err
	}
	if err := keyring.Set(keyringService, k.Profile, string(raw)); err == nil {
		return "keyring", nil
	}
	fp := k.FilePath
	if fp == "" {
		fp = defaultFilePath(k.Profile)
	}
	if err := os.MkdirAll(filepath.Dir(fp), 0o700); err != nil {
		return "", err
	}
	if err := os.WriteFile(fp, raw, 0o600); err != nil {
		return "", err
	}
	return fp, nil
}

// Delete removes the credential from both keyring and fallback file.
func (k KeyringSource) Delete() error {
	_ = keyring.Delete(keyringService, k.Profile)
	fp := k.FilePath
	if fp == "" {
		fp = defaultFilePath(k.Profile)
	}
	if err := os.Remove(fp); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func tokenFromRaw(raw string) (string, error) {
	if raw == "" {
		return "", ErrNotFound
	}
	var cred Credential
	if err := json.Unmarshal([]byte(raw), &cred); err != nil {
		// Treat bare-string legacy value as a personal token.
		return raw, nil
	}
	if cred.Token == "" {
		return "", ErrNotFound
	}
	return cred.Token, nil
}

func defaultFilePath(profile string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "fiken", "credentials", profile+".json")
}
