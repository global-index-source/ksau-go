package crypto

import "github.com/ProtonMail/gopenpgp/v3/crypto"

var pgp *crypto.PGPHandle = crypto.PGP()

func getPrivateKey() *crypto.Key {
	key, err := crypto.NewPrivateKeyFromArmored(privkey, []byte(passphrase))
	if err != nil {
		panic("Failed to create private key")
	}

	return key
}

func Encrypt(text string) []byte {
	encryptionHandler, err := pgp.Encryption().Recipient(getPrivateKey()).New()
	if err != nil {
		panic("Failed to create encryption handler")
	}

	encrypted, err := encryptionHandler.Encrypt([]byte(text))
	if err != nil {
		panic("Failed to encrypt text")
	}

	armorbytes, err := encrypted.ArmorBytes()
	if err != nil {
		panic("Failed to armor bytes")
	}
	return armorbytes
}

func Decrypt(data []byte) []byte {
	decryptionHandler, err := pgp.Decryption().DecryptionKey(getPrivateKey()).New()
	if err != nil {
		panic("Failed to create decryption handler")
	}

	decrypted, err := decryptionHandler.Decrypt(data, crypto.Armor)
	if err != nil {
		panic("Failed to decrypt data")
	}

	return decrypted.Bytes()
}
