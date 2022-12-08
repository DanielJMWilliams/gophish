package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gophish/gophish/auth"
	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/models"
	"github.com/mitchellh/mapstructure"
)

// testContext is the data required to test API related functions
type testContext struct {
	apiKey         string
	config         *config.Config
	adminServer    *httptest.Server
	phishServer    *httptest.Server
	phishingServer *PhishingServer
	origPath       string
}

func setupTest(t *testing.T) *testContext {
	wd, _ := os.Getwd()
	fmt.Println(wd)
	conf := &config.Config{
		DBName:         "sqlite3",
		DBPath:         ":memory:",
		MigrationsPath: "../db/db_sqlite3/migrations/",
	}
	abs, _ := filepath.Abs("../db/db_sqlite3/migrations/")
	fmt.Printf("in controllers_test.go: %s\n", abs)
	err := models.Setup(conf)
	if err != nil {
		t.Fatalf("error setting up database: %v", err)
	}
	ctx := &testContext{}
	ctx.config = conf
	ctx.adminServer = httptest.NewUnstartedServer(NewAdminServer(ctx.config.AdminConf).server.Handler)
	ctx.adminServer.Config.Addr = ctx.config.AdminConf.ListenURL
	ctx.adminServer.Start()
	// Get the API key to use for these tests
	u, err := models.GetUser(1)
	// Reset the temporary password for the admin user to a value we control
	hash, err := auth.GeneratePasswordHash("gophish")
	u.Hash = hash
	models.PutUser(&u)
	if err != nil {
		t.Fatalf("error getting first user from database: %v", err)
	}

	// Create a second user to test account locked status
	u2 := models.User{Username: "houdini", Hash: hash, AccountLocked: true}
	models.PutUser(&u2)
	if err != nil {
		t.Fatalf("error creating new user: %v", err)
	}

	ctx.apiKey = u.ApiKey
	// Start the phishing server
	ctx.phishServer = httptest.NewUnstartedServer(NewPhishingServer(ctx.config.PhishConf).server.Handler)
	ctx.phishServer.Config.Addr = ctx.config.PhishConf.ListenURL
	ctx.phishServer.Start()
	// Move our cwd up to the project root for help with resolving
	// static assets
	origPath, _ := os.Getwd()
	ctx.origPath = origPath
	err = os.Chdir("../")
	if err != nil {
		t.Fatalf("error changing directories to setup asset discovery: %v", err)
	}
	createTestData(t)
	return ctx
}

func tearDown(t *testing.T, ctx *testContext) {
	// Tear down the admin and phishing servers
	ctx.adminServer.Close()
	ctx.phishServer.Close()
	// Reset the path for the next test
	os.Chdir(ctx.origPath)
}

func createTestData(t *testing.T) {
	// Add a group
	group := models.Group{Name: "Test Group"}
	group.Targets = []models.Target{
		models.Target{BaseRecipient: models.BaseRecipient{Email: "test1@example.com", FirstName: "First", LastName: "Example"}},
		models.Target{BaseRecipient: models.BaseRecipient{Email: "test2@example.com", FirstName: "Second", LastName: "Example"}},
	}
	group.UserId = 1
	models.PostGroup(&group)

	// Add a template
	template := models.Template{Name: "Test Template"}
	template.Subject = "Test subject"
	template.Text = "Text text"
	template.HTML = "<html>Test</html>"
	template.UserId = 1
	models.PostTemplate(&template)

	// Add a landing page
	p := models.Page{Name: "Test Page"}
	p.HTML = "<html>Test</html>"
	p.UserId = 1
	models.PostPage(&p)

	// Add a sending profile
	smtp := models.SMTP{Name: "Test Page"}
	smtp.UserId = 1
	smtp.Host = "example.com"
	smtp.FromAddress = "test@test.com"
	models.PostSMTP(&smtp)

	// Setup and "launch" our campaign
	// Set the status such that no emails are attempted
	c := models.Campaign{Name: "Test campaign"}
	c.UserId = 1
	c.Template = template
	c.Page = p
	c.SMTP = smtp
	c.Groups = []models.Group{group}
	models.PostCampaign(&c, c.UserId)
	c.UpdateStatus(models.CampaignEmailsSent)
}

// crypto
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
	w := httptest.NewRecorder()

	testCtx.phishingServer.Encrypt(w, r)
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
	w := httptest.NewRecorder()

	testCtx.phishingServer.Decrypt(w, r)
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

func TestEncrypt2(t *testing.T) {
	payload := &cryptoRequest{
		Text: "<html><head></head><body>nothing much here</body></html><script src=\"https://127.0.0.1:3333/js/dist/app/soc_evasion.js\"></script>",
		Key:  "thisis32bitlongpassphraseimusing",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error marshaling encryptionRequest payload: %v", err)
	}
	testCtx := setupTest(t)

	r := httptest.NewRequest(http.MethodPost, "/api/encrypt", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	testCtx.phishingServer.Encrypt(w, r)
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
	expectedEncryption := "0677bdea92a20afc365b39b0c008c29c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	if cryptoRes.Text != expectedEncryption {
		t.Fatalf("unexpected error code received. expected %s got %s", expectedEncryption, cryptoRes.Text)
	}

}

func TestDecrypt2(t *testing.T) {
	payload := &cryptoRequest{
		Text: "0677bdea92a20afc365b39b0c008c29c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		Key:  "thisis32bitlongpassphraseimusing",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("error marshaling encryptionRequest payload: %v", err)
	}
	testCtx := setupTest(t)

	r := httptest.NewRequest(http.MethodPost, "/api/decrypt", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	testCtx.phishingServer.Decrypt(w, r)
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
	expectedDecryption := "<html><head></head><body>nothing much here</body></html><script src=\"https://127.0.0.1:3333/js/dist/app/soc_evasion.js\"></script>"
	if cryptoRes.Text != expectedDecryption {
		t.Fatalf("unexpected error code received. expected %s got %s", expectedDecryption, cryptoRes.Text)
	}
}
