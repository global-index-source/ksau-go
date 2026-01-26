package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/global-index-source/ksau-go/azure"
	"github.com/global-index-source/ksau-go/internal/config"
	"github.com/global-index-source/ksau-go/internal/rcloneconfig"
	"github.com/spf13/cobra"
)

var (
	remoteName       string
	remoteType       string
	remoteTokenFile  string
	remoteClientID   string
	remoteSecret     string
	remoteDriveID    string
	remoteDriveType  string
	remoteRootFolder string
	remoteBaseURL    string
	remoteDryRun     bool
	remoteForce      bool
	remoteShowSecret bool
	remoteExportFile string
	remoteImportFile string
	remoteLiveCheck  bool
)

var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Manage remote configurations",
}

var remoteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available remotes",
	Run:   runRemoteList,
}

var remoteAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new remote",
	Run:   runRemoteAdd,
}

var remoteUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing remote",
	Run:   runRemoteUpdate,
}

var remoteRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a remote",
	Run:   runRemoteRemove,
}

var remoteValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a remote configuration",
	Run:   runRemoteValidate,
}

var remoteExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export remotes to JSON",
	Run:   runRemoteExport,
}

var remoteImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import remotes from JSON",
	Run:   runRemoteImport,
}

func init() {
	rootCmd.AddCommand(remoteCmd)
	remoteCmd.AddCommand(remoteListCmd)
	remoteCmd.AddCommand(remoteAddCmd)
	remoteCmd.AddCommand(remoteUpdateCmd)
	remoteCmd.AddCommand(remoteRemoveCmd)
	remoteCmd.AddCommand(remoteValidateCmd)
	remoteCmd.AddCommand(remoteExportCmd)
	remoteCmd.AddCommand(remoteImportCmd)

	remoteAddCmd.Flags().StringVar(&remoteName, "name", "", "Remote name (required)")
	remoteAddCmd.Flags().StringVar(&remoteType, "type", "onedrive", "Remote type")
	remoteAddCmd.Flags().StringVar(&remoteTokenFile, "token-file", "", "Path to token JSON file")
	remoteAddCmd.Flags().StringVar(&remoteClientID, "client-id", "", "Client ID")
	remoteAddCmd.Flags().StringVar(&remoteSecret, "client-secret", "", "Client secret")
	remoteAddCmd.Flags().StringVar(&remoteDriveID, "drive-id", "", "Drive ID")
	remoteAddCmd.Flags().StringVar(&remoteDriveType, "drive-type", "", "Drive type")
	remoteAddCmd.Flags().StringVar(&remoteRootFolder, "root-folder", "", "Remote root folder")
	remoteAddCmd.Flags().StringVar(&remoteBaseURL, "base-url", "", "Remote base URL")
	remoteAddCmd.Flags().BoolVar(&remoteDryRun, "dry-run", false, "Show changes without writing")
	remoteAddCmd.Flags().BoolVar(&remoteForce, "force", false, "Overwrite existing remote")
	remoteAddCmd.MarkFlagRequired("name")
	remoteAddCmd.MarkFlagRequired("token-file")

	remoteUpdateCmd.Flags().StringVar(&remoteName, "name", "", "Remote name (required)")
	remoteUpdateCmd.Flags().StringVar(&remoteType, "type", "", "Remote type")
	remoteUpdateCmd.Flags().StringVar(&remoteTokenFile, "token-file", "", "Path to token JSON file")
	remoteUpdateCmd.Flags().StringVar(&remoteClientID, "client-id", "", "Client ID")
	remoteUpdateCmd.Flags().StringVar(&remoteSecret, "client-secret", "", "Client secret")
	remoteUpdateCmd.Flags().StringVar(&remoteDriveID, "drive-id", "", "Drive ID")
	remoteUpdateCmd.Flags().StringVar(&remoteDriveType, "drive-type", "", "Drive type")
	remoteUpdateCmd.Flags().StringVar(&remoteRootFolder, "root-folder", "", "Remote root folder")
	remoteUpdateCmd.Flags().StringVar(&remoteBaseURL, "base-url", "", "Remote base URL")
	remoteUpdateCmd.Flags().BoolVar(&remoteDryRun, "dry-run", false, "Show changes without writing")
	remoteUpdateCmd.MarkFlagRequired("name")

	remoteRemoveCmd.Flags().StringVar(&remoteName, "name", "", "Remote name (required)")
	remoteRemoveCmd.Flags().BoolVar(&remoteDryRun, "dry-run", false, "Show changes without writing")
	remoteRemoveCmd.MarkFlagRequired("name")

	remoteValidateCmd.Flags().StringVar(&remoteName, "name", "", "Remote name (required)")
	remoteValidateCmd.Flags().BoolVar(&remoteLiveCheck, "live", false, "Perform live validation using Graph API")
	remoteValidateCmd.MarkFlagRequired("name")

	remoteListCmd.Flags().BoolVar(&remoteShowSecret, "show-secrets", false, "Show tokens in output")
	remoteExportCmd.Flags().BoolVar(&remoteShowSecret, "show-secrets", false, "Show tokens in export")
	remoteExportCmd.Flags().StringVar(&remoteExportFile, "file", "", "Write output to file (defaults to stdout)")
	remoteImportCmd.Flags().StringVar(&remoteImportFile, "file", "", "Input JSON file (required)")
	remoteImportCmd.Flags().BoolVar(&remoteDryRun, "dry-run", false, "Show changes without writing")
	remoteImportCmd.Flags().BoolVar(&remoteForce, "force", false, "Overwrite existing remotes")
	remoteImportCmd.MarkFlagRequired("file")
}

type remoteExportEntry struct {
	Name string            `json:"name"`
	Keys map[string]string `json:"keys"`
}

func runRemoteList(cmd *cobra.Command, args []string) {
	cfg, _, err := loadRcloneConfig()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(1)
	}

	if len(cfg.Sections) == 0 {
		fmt.Println("No remotes found.")
		return
	}

	fmt.Println("Remotes:")
	for _, name := range cfg.Order {
		section := cfg.Sections[name]
		line := fmt.Sprintf("- %s", name)
		if section != nil {
			if value := section.Keys["type"]; value != "" {
				line += fmt.Sprintf(" (type=%s)", value)
			}
			if remoteShowSecret {
				if token := section.Keys["token"]; token != "" {
					line += fmt.Sprintf(" token=%s", token)
				}
			}
		}
		fmt.Println(line)
	}
}

func runRemoteAdd(cmd *cobra.Command, args []string) {
	cfg, existingFormat, err := loadRcloneConfig()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(1)
	}

	if _, exists := cfg.Sections[remoteName]; exists && !remoteForce {
		fmt.Println("Remote already exists. Use --force to overwrite.")
		os.Exit(1)
	}

	section := cfg.EnsureSection(remoteName)
	section.Set("type", remoteType)
	applyRemoteFields(section)

	if err := writeRcloneConfig(cfg, existingFormat, remoteDryRun); err != nil {
		fmt.Println("Failed to write config:", err)
		os.Exit(1)
	}

	if remoteDryRun {
		fmt.Println("Dry run complete. No changes written.")
		return
	}

	fmt.Println("Remote added successfully.")
}

func runRemoteUpdate(cmd *cobra.Command, args []string) {
	cfg, existingFormat, err := loadRcloneConfig()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(1)
	}

	section, exists := cfg.Sections[remoteName]
	if !exists {
		fmt.Println("Remote not found:", remoteName)
		os.Exit(1)
	}

	if remoteType != "" {
		section.Set("type", remoteType)
	}
	applyRemoteFields(section)

	if err := writeRcloneConfig(cfg, existingFormat, remoteDryRun); err != nil {
		fmt.Println("Failed to write config:", err)
		os.Exit(1)
	}

	if remoteDryRun {
		fmt.Println("Dry run complete. No changes written.")
		return
	}

	fmt.Println("Remote updated successfully.")
}

func runRemoteRemove(cmd *cobra.Command, args []string) {
	cfg, existingFormat, err := loadRcloneConfig()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(1)
	}

	if removed := cfg.RemoveSection(remoteName); !removed {
		fmt.Println("Remote not found:", remoteName)
		os.Exit(1)
	}

	if err := writeRcloneConfig(cfg, existingFormat, remoteDryRun); err != nil {
		fmt.Println("Failed to write config:", err)
		os.Exit(1)
	}

	if remoteDryRun {
		fmt.Println("Dry run complete. No changes written.")
		return
	}

	fmt.Println("Remote removed successfully.")
}

func runRemoteValidate(cmd *cobra.Command, args []string) {
	cfg, _, err := loadRcloneConfig()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(1)
	}

	section, ok := cfg.Sections[remoteName]
	if !ok {
		fmt.Println("Remote not found:", remoteName)
		os.Exit(1)
	}

	missing := requiredMissing(section)
	if len(missing) > 0 {
		fmt.Println("Remote configuration missing fields:", strings.Join(missing, ", "))
		os.Exit(1)
	}

	if !remoteLiveCheck {
		fmt.Println("Remote configuration looks valid.")
		return
	}

	configData, _, err := getConfigData()
	if err != nil {
		fmt.Println("Failed to load config data:", err)
		os.Exit(1)
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	client, err := azure.NewAzureClientFromRcloneConfigData(configData, remoteName)
	if err != nil {
		fmt.Println("Failed to initialize remote:", err)
		os.Exit(1)
	}

	_, err = client.GetDriveQuota(httpClient)
	if err != nil {
		fmt.Println("Live validation failed:", err)
		os.Exit(1)
	}

	fmt.Println("Remote validated successfully.")
}

func runRemoteExport(cmd *cobra.Command, args []string) {
	cfg, _, err := loadRcloneConfig()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(1)
	}

	entries := make([]remoteExportEntry, 0, len(cfg.Sections))
	for _, name := range cfg.Order {
		section := cfg.Sections[name]
		if section == nil {
			continue
		}
		keys := make(map[string]string, len(section.Keys))
		for key, value := range section.Keys {
			if key == "token" && !remoteShowSecret {
				keys[key] = "<redacted>"
				continue
			}
			keys[key] = value
		}
		entries = append(entries, remoteExportEntry{Name: name, Keys: keys})
	}

	payload, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		fmt.Println("Failed to encode export:", err)
		os.Exit(1)
	}

	if remoteExportFile == "" {
		fmt.Println(string(payload))
		return
	}

	if err := os.WriteFile(remoteExportFile, payload, 0600); err != nil {
		fmt.Println("Failed to write export file:", err)
		os.Exit(1)
	}
	fmt.Println("Export written to", remoteExportFile)
}

func runRemoteImport(cmd *cobra.Command, args []string) {
	payload, err := os.ReadFile(remoteImportFile)
	if err != nil {
		fmt.Println("Failed to read import file:", err)
		os.Exit(1)
	}

	var entries []remoteExportEntry
	if err := json.Unmarshal(payload, &entries); err != nil {
		fmt.Println("Failed to parse import file:", err)
		os.Exit(1)
	}

	cfg, existingFormat, err := loadRcloneConfig()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.Name == "" {
			fmt.Println("Skipping entry with empty name")
			continue
		}
		if _, exists := cfg.Sections[entry.Name]; exists && !remoteForce {
			fmt.Printf("Remote '%s' exists. Use --force to overwrite.\n", entry.Name)
			continue
		}
		section := cfg.EnsureSection(entry.Name)
		section.Keys = make(map[string]string)
		section.KeyOrder = nil
		for key, value := range entry.Keys {
			if key == "token" && value == "<redacted>" {
				continue
			}
			section.Set(key, value)
		}
	}

	if err := writeRcloneConfig(cfg, existingFormat, remoteDryRun); err != nil {
		fmt.Println("Failed to write config:", err)
		os.Exit(1)
	}

	if remoteDryRun {
		fmt.Println("Dry run complete. No changes written.")
		return
	}

	fmt.Println("Import completed successfully.")
}

func applyRemoteFields(section *rcloneconfig.Section) {
	if remoteTokenFile != "" {
		payload, err := os.ReadFile(remoteTokenFile)
		if err != nil {
			fmt.Println("Failed to read token file:", err)
			os.Exit(1)
		}
		section.Set("token", strings.TrimSpace(string(payload)))
	}

	if remoteClientID != "" {
		section.Set("client_id", remoteClientID)
	}
	if remoteSecret != "" {
		section.Set("client_secret", remoteSecret)
	}
	if remoteDriveID != "" {
		section.Set("drive_id", remoteDriveID)
	}
	if remoteDriveType != "" {
		section.Set("drive_type", remoteDriveType)
	}
	if remoteRootFolder != "" {
		section.Set("root_folder", remoteRootFolder)
	}
	if remoteBaseURL != "" {
		section.Set("base_url", remoteBaseURL)
	}
}

func requiredMissing(section *rcloneconfig.Section) []string {
	requiredKeys := []string{"token", "client_id", "client_secret", "drive_id", "drive_type"}
	var missing []string
	for _, key := range requiredKeys {
		if strings.TrimSpace(section.Keys[key]) == "" {
			missing = append(missing, key)
		}
	}
	return missing
}

func loadRcloneConfig() (*rcloneconfig.Config, *config.Format, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, nil, err
	}

	rawData, format, err := config.ReadConfigFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return rcloneconfig.NewConfig(), nil, nil
		}
		return nil, nil, err
	}

	decoded, err := config.DecodeConfig(rawData, format)
	if err != nil {
		return nil, nil, err
	}

	parsed, err := rcloneconfig.Parse(decoded)
	if err != nil {
		return nil, nil, err
	}

	return parsed, &format, nil
}

func writeRcloneConfig(cfg *rcloneconfig.Config, existingFormat *config.Format, dryRun bool) error {
	writeFormat, err := config.ResolveWriteFormat(configFormat, existingFormat)
	if err != nil {
		return err
	}

	plaintext := cfg.Render()
	encoded, err := config.EncodeConfig(plaintext, writeFormat)
	if err != nil {
		return err
	}

	if dryRun {
		return nil
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	return config.WriteConfigFile(configPath, encoded)
}
