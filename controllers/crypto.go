package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gophish/gophish/controllers/api"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/util/crypto"
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
	encryptedMessage := crypto.EncryptGCM(p.Text, []byte(p.Key))
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
	decryptedMessage := crypto.DecryptGCM(p.Text, []byte(p.Key))
	log.Info("DECRYPTED MESSAGE: ")
	log.Info(decryptedMessage)
	res := cryptoResponse{Text: decryptedMessage}
	api.JSONResponse(w, models.Response{Success: true, Message: "Text encrypted successfully", Data: res}, http.StatusOK)
}
