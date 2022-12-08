package controllers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gophish/gophish/controllers/api"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
)

//CRYPTO

type cryptoResponse struct {
	Text string `json:"text"`
}

type cryptoRequest struct {
	Text string `json:"text"`
	Key  string `json:"key"`
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	//(*w).Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
	log.Info("CORS enabled")
}

func (ps *PhishingServer) Encrypt(w http.ResponseWriter, r *http.Request) {
	log.Info("ENCRYPT")
	enableCors(&w)
	if r.Method != "POST" {
		api.JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusBadRequest)
		return
	}
	p := cryptoRequest{}
	// Put the request into a page
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		api.JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
		return
	}
	//encryptedMessage := EncryptAES([]byte(p.Key), p.Text)
	encryptedMessage := EncryptGCM(p.Text, []byte(p.Key))
	res := cryptoResponse{Text: encryptedMessage}
	api.JSONResponse(w, models.Response{Success: true, Message: "Text encrypted successfully", Data: res}, http.StatusOK)
}

func (ps *PhishingServer) Decrypt(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	//JSONResponse(w, models.Response{Success: false, Message: "API CALLED SUCCESSFULLY"}, http.StatusBadRequest)
	if r.Method != "POST" {
		api.JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusBadRequest)
		return
	}
	p := cryptoRequest{}
	// Put the request into a page
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		api.JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
		return
	}
	//decryptedMessage := DecryptAES([]byte(p.Key), p.Text)
	decryptedMessage := DecryptGCM(p.Text, []byte(p.Key))
	log.Info("DECRYPTED MESSAGE: ")
	log.Info(decryptedMessage)
	res := cryptoResponse{Text: decryptedMessage}
	api.JSONResponse(w, models.Response{Success: true, Message: "Text encrypted successfully", Data: res}, http.StatusOK)
}

func EncryptAES(key []byte, plaintext string) string {
	// create cipher
	c, _ := aes.NewCipher(key)

	// allocate space for ciphered data
	out := make([]byte, len(plaintext))

	// encrypt
	c.Encrypt(out, []byte(plaintext))
	// return hex string
	return hex.EncodeToString(out)
}

func DecryptAES(key []byte, ct string) string {
	ciphertext, _ := hex.DecodeString(ct)

	c, _ := aes.NewCipher(key)

	pt := make([]byte, len(ciphertext))
	c.Decrypt(pt, ciphertext)

	s := string(pt[:])
	return s
}

func EncryptGCM(stringToEncrypt string, keyString []byte) (encryptedString string) {

	//Since the key is in string, we need to convert decode it to bytes
	key := keyString
	plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext)
}

func DecryptGCM(encryptedString string, keyString []byte) (decryptedString string) {

	key := keyString
	enc, _ := hex.DecodeString(encryptedString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return fmt.Sprintf("%s", plaintext)
}
