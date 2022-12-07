package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gophish/gophish/models"
	"github.com/mitchellh/mapstructure"
)

func TestEncrypt(t *testing.T) {
	payload := &cryptoRequest{
		Text: "This is a secret",
		Key:  "thisis32bitlongpassphraseimusing",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error marshaling encryptionRequest payload: %v", err)
	}
	testCtx := setupTest(t)

	r := httptest.NewRequest(http.MethodPost, "/api/encrypt", bytes.NewBuffer(body))
	r = ctx1.Set(r, "user", testCtx.admin)
	w := httptest.NewRecorder()

	testCtx.apiServer.Encrypt(w, r)
	expected := http.StatusOK
	if w.Code != expected {
		t.Fatalf("unexpected error code received. expected %d got %d", expected, w.Code)
	}
	resBytes := w.Body.Bytes()

	apiRes := models.Response{}
	if err := json.Unmarshal(resBytes, &apiRes); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	var cryptoRes cryptoResponse
	mapstructure.Decode(apiRes.Data, &cryptoRes)
	//check correct encryption
	expectedEncryption := "145149d64a1a3c4025e67665001a3167"
	if cryptoRes.Text != expectedEncryption {
		t.Fatalf("unexpected error code received. expected %s got %s", expectedEncryption, cryptoRes.Text)
	}

}

func TestDecrypt(t *testing.T) {
	payload := &cryptoRequest{
		Text: "145149d64a1a3c4025e67665001a3167",
		Key:  "thisis32bitlongpassphraseimusing",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error marshaling encryptionRequest payload: %v", err)
	}
	testCtx := setupTest(t)

	r := httptest.NewRequest(http.MethodPost, "/api/decrypt", bytes.NewBuffer(body))
	r = ctx1.Set(r, "user", testCtx.admin)
	w := httptest.NewRecorder()

	testCtx.apiServer.Decrypt(w, r)
	expected := http.StatusOK
	if w.Code != expected {
		t.Fatalf("unexpected error code received. expected %d got %d", expected, w.Code)
	}
	resBytes := w.Body.Bytes()

	apiRes := models.Response{}
	if err := json.Unmarshal(resBytes, &apiRes); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	var cryptoRes cryptoResponse
	mapstructure.Decode(apiRes.Data, &cryptoRes)
	//check correct decryption
	expectedDecryption := "This is a secret"
	if cryptoRes.Text != expectedDecryption {
		t.Fatalf("unexpected error code received. expected %s got %s", expectedDecryption, cryptoRes.Text)
	}

}
