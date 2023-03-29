package crypto

import (
	"testing"
)

// crypto
func TestEncryptAndDecrypt(t *testing.T) {
	plaintext := "This is a secret"
	key := "11111111112222222222333333333312"

	ciphertext := EncryptGCM(plaintext, []byte(key))

	decryptedMessage := DecryptGCM(ciphertext, []byte(key))

	if plaintext != decryptedMessage {
		t.Fatalf("unexpected error code received. expected %s got %s", plaintext, decryptedMessage)
	}

}

func TestEncryptAndDecrypt2(t *testing.T) {
	plaintext := "<html><head></head><body>nothing much here</body></html><script src=\"http://127.0.0.1:80/static/proxy_bypass.js\"></script>"
	key := "noZsmWPQZzGVJzkjtiTvVzkPQhI9MwMM"

	ciphertext := EncryptGCM(plaintext, []byte(key))

	decryptedMessage := DecryptGCM(ciphertext, []byte(key))

	if plaintext != decryptedMessage {
		t.Fatalf("unexpected error code received. expected %s got %s", plaintext, decryptedMessage)
	}

}
