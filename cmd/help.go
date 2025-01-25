package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Get detailed help about commands",
	Long: `Get detailed help and usage examples for ksau-go commands.
If no command is specified, displays help about all commands.`,
	Run: runHelp,
}

func init() {
	rootCmd.AddCommand(helpCmd)
}

func runHelp(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("ksau-go - OneDrive Upload Utility")
		fmt.Println("\nAvailable Commands:")
		fmt.Println("\nupload - Upload files to OneDrive")
		fmt.Println("  Examples:")
		fmt.Println("    # Upload a file to the root folder")
		fmt.Println("    ksau-go upload -f myfile.txt -r /")
		fmt.Println("    # Upload with custom remote name")
		fmt.Println("    ksau-go upload -f local.txt -r /docs -n remote.txt")
		fmt.Println("    # Upload with specific chunk size (in bytes)")
		fmt.Println("    ksau-go upload -f large.zip -r /backup -s 8388608")
		fmt.Println("    # Upload using different remote config")
		fmt.Println("    ksau-go upload -f file.pdf -r /shared --remote-config saurajcf")

		fmt.Println("\nquota - Display OneDrive quota information")
		fmt.Println("  Examples:")
		fmt.Println("    # Show quota for all remotes")
		fmt.Println("    ksau-go quota")
		fmt.Println("    # Show quota for specific remote")
		fmt.Println("    ksau-go quota --remote-config oned")

		fmt.Println("\nversion - Show version information")
		fmt.Println("  Example:")
		fmt.Println("    ksau-go version")

		fmt.Println("\nGlobal Flags:")
		fmt.Println("  --remote-config  Name of the remote configuration (default: oned)")
	} else {
		fmt.Printf("Help for '%s' command:\n", args[0])
		switch args[0] {
		case "upload":
			printUploadHelp()
		case "quota":
			printQuotaHelp()
		case "version":
			printVersionHelp()
		case "refresh":
			printRefreshHelp()
		case "list-remote":
			printListRemoteHelp()
		default:
			fmt.Printf("Unknown command: %s\n", args[0])
		}
	}
}

func printUploadHelp() {
	fmt.Println(`
Upload Command
-------------
Upload files to OneDrive with support for chunked uploads and integrity verification.

Usage:
  ksau-go upload -f <file> -r <remote-path> [flags]

Required Flags:
  -f, --file          Path to the local file to upload
  -r, --remote        Remote folder path on OneDrive

Optional Flags:
  -n, --remote-name     Custom name for the uploaded file
  -s, --chunk-size      Size of upload chunks in bytes (default: automatic)
  -p, --parallel        Number of parallel upload chunks (default: 1)
      --retries         Maximum upload retry attempts (default: 3)
      --retry-delay     Delay between retries (default: 5s)
      --skip-hash       Skip file integrity verification
      --hash-retries    Maximum hash verification retries (default: 5)

Examples:
  # Basic file upload
  ksau-go upload -f document.pdf -r /Documents

  # Upload with different name
  ksau-go upload -f local.txt -r /Backup -n remote.txt

  # Upload large file with custom chunk size
  ksau-go upload -f large.iso -r /ISOs -s 16777216 -p 4`)
}

func printQuotaHelp() {
	fmt.Println(`
Quota Command
------------
Display storage quota information for OneDrive remotes.

Usage:
  ksau-go quota [flags]

The quota command will display:
- Total space
- Used space
- Available space
- Usage percentage

For each configured remote (oned, saurajcf, etc.)

Example:
  ksau-go quota`)
}

func printVersionHelp() {
	fmt.Println(`
Version Command
--------------
Display version information for ksau-go.

Usage:
  ksau-go version

Shows:
- Version number
- Commit hash
- Build date

Example:
  ksau-go version`)
}

func printRefreshHelp() {
	fmt.Println(`
Refresh Command
---------------
Refresh the configuration file and cache.

Usage:
  ksau-go refresh [flags]
  
Optional Flags:
  -u, --url     Custom URL to fetch the configuration file (must be direct).

Note:
  The configuration file is encrypted and stored in common config path for your OS.
  It is decrypted in memory, so there is no point trying to read it yourself.`)
}

func printListRemoteHelp() {
	fmt.Println(`
List Remote Command
-------------------
List available remotes from the configuration file.

Usage:
  ksau-go list-remote

Note:
  This command will list all available remotes from the configuration file.
  If the command fails, run refresh.`)
}
