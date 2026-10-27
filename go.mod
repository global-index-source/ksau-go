module github.com/global-index-source/ksau-go

go 1.24.0

// Builds from source pick up a toolchain that distrusts the compromised
// eMudhra intermediates (GO-2026-6809).
toolchain go1.27.3

require (
	github.com/ProtonMail/gopenpgp/v3 v3.1.2
	github.com/spf13/cobra v1.8.1
)

require (
	github.com/ProtonMail/go-crypto v1.1.5 // indirect
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
)

replace github.com/rclone/rclone => github.com/rclone/rclone v1.65.2
