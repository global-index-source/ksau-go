package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/global-index-source/ksau-go/crypto"
	"github.com/global-index-source/ksau-go/internal/rcloneconfig"
)

type Format string

type FormatPreference string

const (
	FormatEncrypted Format = "encrypted"
	FormatPlaintext Format = "plaintext"

	FormatAuto FormatPreference = "auto"
)

func DetectFormat(data []byte) (Format, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return "", errors.New("config file is empty")
	}
	if bytes.Contains(trimmed, []byte("-----BEGIN PGP MESSAGE-----")) {
		return FormatEncrypted, nil
	}
	if _, err := rcloneconfig.Parse(trimmed); err != nil {
		return "", fmt.Errorf("unrecognized config format: %w", err)
	}
	return FormatPlaintext, nil
}

func ReadConfigFile(path string) ([]byte, Format, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	format, err := DetectFormat(data)
	if err != nil {
		return nil, "", err
	}
	return data, format, nil
}

func DecodeConfig(data []byte, format Format) ([]byte, error) {
	switch format {
	case FormatPlaintext:
		return data, nil
	case FormatEncrypted:
		return crypto.Decrypt(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func EncodeConfig(data []byte, format Format) ([]byte, error) {
	switch format {
	case FormatPlaintext:
		return data, nil
	case FormatEncrypted:
		return crypto.Encrypt(string(data))
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func CanEncrypt() bool {
	_, err := crypto.Encrypt("ping")
	return err == nil
}

func ResolveWriteFormat(preference string, existing *Format) (Format, error) {
	if preference == "" {
		preference = string(FormatAuto)
	}
	switch FormatPreference(strings.ToLower(preference)) {
	case FormatAuto:
		if existing != nil {
			return *existing, nil
		}
		if CanEncrypt() {
			return FormatEncrypted, nil
		}
		return FormatPlaintext, nil
	case FormatPreference(FormatEncrypted):
		if !CanEncrypt() {
			return "", errors.New("encryption requested but key material is unavailable")
		}
		return FormatEncrypted, nil
	case FormatPreference(FormatPlaintext):
		return FormatPlaintext, nil
	default:
		return "", fmt.Errorf("invalid config format: %s", preference)
	}
}

func WriteConfigFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if _, err := os.Stat(path); err == nil {
		backupPath := fmt.Sprintf("%s.bak-%s", path, time.Now().UTC().Format("20060102-150405"))
		if err := copyFile(path, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	tmpFile, err := os.CreateTemp(dir, filepath.Base(path)+".tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), path); err != nil {
		return fmt.Errorf("failed to replace config file: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(path, 0600); err != nil {
			return fmt.Errorf("failed to set config permissions: %w", err)
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}
