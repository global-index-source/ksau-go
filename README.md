# ksau-go

## Development
> Table of Content
> - [Building the Project](#Building-the-Project)

### Building the Project
To build `ksau-go`, there's one important detail to note. You will need to create
**crypto/credentials.go**. Then, set constants named `privkey` which is a secret pgp
key, and `passphrase` which is the passphrase for that secret pgp key (if any),
else set an empty const.

Follow this template:
```go
//go:build credentials
package crypto

const privkey string = `mykey`
const passphrase string = "mypassphrase"
```
> [!NOTE]
> The `//go:build credentials` directive is **not** optional. do **not** omit it.

Then, run the following to build:
```bash
go build -tags credentials
```

