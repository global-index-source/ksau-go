package crypto

import (
	"fmt"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

var pgp *crypto.PGPHandle = crypto.PGP()

func getPrivateKey() *crypto.Key {
	key, err := crypto.NewPrivateKeyFromArmored(privkey, []byte(passphrase))
	if err != nil {
		panic("Failed to create private key")
	}

	return key
}

func Encrypt(text string) ([]byte, error) {
	encryptionHandler, err := pgp.Encryption().Recipient(getPrivateKey()).New()
	if err != nil {
		return nil, fmt.Errorf("failed to create encryption handler: %w", err)
	}

	encrypted, err := encryptionHandler.Encrypt([]byte(text))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt text: %w", err)
	}

	armorbytes, err := encrypted.ArmorBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to armor bytes: %w", err)
	}
	return armorbytes, nil
}

func Decrypt(data []byte) ([]byte, error) {
	decryptionHandler, err := pgp.Decryption().DecryptionKey(getPrivateKey()).New()
	if err != nil {
		return nil, fmt.Errorf("failed to create decryption handler: %w", err)
	}

	decrypted, err := decryptionHandler.Decrypt(data, crypto.Armor)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}
	return decrypted.Bytes(), nil
}
