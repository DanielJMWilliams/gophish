package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gophish/gophish/config"
	ctx1 "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
	"github.com/mitchellh/mapstructure"
)

type testContext struct {
	apiKey    string
	config    *config.Config
	apiServer *Server
	admin     models.User
}

func setupTest(t *testing.T) *testContext {
	conf := &config.Config{
		DBName:         "sqlite3",
		DBPath:         ":memory:",
		MigrationsPath: "../../db/db_sqlite3/migrations/",
	}
	err := models.Setup(conf)
	if err != nil {
		t.Fatalf("Failed creating database: %v", err)
	}
	ctx := &testContext{}
	ctx.config = conf
	// Get the API key to use for these tests
	u, err := models.GetUser(1)
	if err != nil {
		t.Fatalf("error getting admin user: %v", err)
	}
	ctx.apiKey = u.ApiKey
	ctx.admin = u
	ctx.apiServer = NewServer()
	return ctx
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

func TestSiteImportBaseHref(t *testing.T) {
	ctx := setupTest(t)
	h := "<html><head></head><body><img src=\"/test.png\"/></body></html>"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, h)
	}))
	expected := fmt.Sprintf("<html><head><base href=\"%s\"/></head><body><img src=\"/test.png\"/>\n</body></html>", ts.URL)
	defer ts.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/import/site",
		bytes.NewBuffer([]byte(fmt.Sprintf(`
			{
				"url" : "%s",
				"include_resources" : false
			}
		`, ts.URL))))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	ctx.apiServer.ImportSite(response, req)
	cs := cloneResponse{}
	err := json.NewDecoder(response.Body).Decode(&cs)
	if err != nil {
		t.Fatalf("error decoding response: %v", err)
	}
	if cs.HTML != expected {
		t.Fatalf("unexpected response received. expected %s got %s", expected, cs.HTML)
	}
}
