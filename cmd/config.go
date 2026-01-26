package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/global-index-source/ksau-go/internal/config"
	"github.com/spf13/cobra"
)

var confirmDecrypt bool

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage local configuration file",
}

var configStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show configuration file status",
	Run:   runConfigStatus,
}

var configEncryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt the local configuration file",
	Run:   runConfigEncrypt,
}

var configDecryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt the local configuration file",
	Run:   runConfigDecrypt,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configStatusCmd)
	configCmd.AddCommand(configEncryptCmd)
	configCmd.AddCommand(configDecryptCmd)
	configDecryptCmd.Flags().BoolVar(&confirmDecrypt, "yes-i-understand", false, "Confirm you want to write a plaintext config file")
}

func runConfigStatus(cmd *cobra.Command, args []string) {
	configPath, err := getConfigPath()
	if err != nil {
		fmt.Println("Failed to resolve config path:", err)
		os.Exit(1)
	}

	info, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Config file not found:", configPath)
			return
		}
		fmt.Println("Failed to stat config file:", err)
		os.Exit(1)
	}

	data, format, err := config.ReadConfigFile(configPath)
	if err != nil {
		fmt.Println("Failed to read config file:", err)
		os.Exit(1)
	}

	_, err = config.DecodeConfig(data, format)
	if err != nil {
		fmt.Println("Config format detected:", format)
		fmt.Println("Config decode error:", err)
		return
	}

	fmt.Println("Config path:", configPath)
	fmt.Println("Format:", format)
	fmt.Println("Last modified:", info.ModTime().Format(time.RFC3339))
}

func runConfigEncrypt(cmd *cobra.Command, args []string) {
	configPath, err := getConfigPath()
	if err != nil {
		fmt.Println("Failed to resolve config path:", err)
		os.Exit(1)
	}

	rawData, format, err := config.ReadConfigFile(configPath)
	if err != nil {
		fmt.Println("Failed to read config file:", err)
		os.Exit(1)
	}

	if format == config.FormatEncrypted {
		fmt.Println("Config file is already encrypted.")
		return
	}

	decoded, err := config.DecodeConfig(rawData, format)
	if err != nil {
		fmt.Println("Failed to decode config:", err)
		os.Exit(1)
	}

	encoded, err := config.EncodeConfig(decoded, config.FormatEncrypted)
	if err != nil {
		fmt.Println("Failed to encrypt config:", err)
		os.Exit(1)
	}

	if err := config.WriteConfigFile(configPath, encoded); err != nil {
		fmt.Println("Failed to write encrypted config:", err)
		os.Exit(1)
	}

	fmt.Println("Config file encrypted successfully.")
}

func runConfigDecrypt(cmd *cobra.Command, args []string) {
	if !confirmDecrypt {
		fmt.Println("Refusing to decrypt without --yes-i-understand")
		os.Exit(1)
	}

	configPath, err := getConfigPath()
	if err != nil {
		fmt.Println("Failed to resolve config path:", err)
		os.Exit(1)
	}

	rawData, format, err := config.ReadConfigFile(configPath)
	if err != nil {
		fmt.Println("Failed to read config file:", err)
		os.Exit(1)
	}

	if format == config.FormatPlaintext {
		fmt.Println("Config file is already plaintext.")
		return
	}

	decoded, err := config.DecodeConfig(rawData, format)
	if err != nil {
		fmt.Println("Failed to decode config:", err)
		os.Exit(1)
	}

	encoded, err := config.EncodeConfig(decoded, config.FormatPlaintext)
	if err != nil {
		fmt.Println("Failed to prepare plaintext config:", err)
		os.Exit(1)
	}

	if err := config.WriteConfigFile(configPath, encoded); err != nil {
		fmt.Println("Failed to write plaintext config:", err)
		os.Exit(1)
	}

	fmt.Println("Config file decrypted successfully.")
}
