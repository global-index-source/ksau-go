package crypto

import _ "embed"

//go:embed passphrase.txt
var passphrase string

//go:embed privkey.pem
var privkey string
