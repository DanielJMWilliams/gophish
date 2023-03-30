package controllers

import (
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gophish/gophish/auth"
	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/models"
)

// testContext is the data required to test API related functions
type testContext struct {
	apiKey      string
	config      *config.Config
	adminServer *httptest.Server
	phishServer *httptest.Server
	origPath    string
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

	//add decoy page
	p2 := models.Page{Name: "Decoy Page"}
	p2.HTML = "<html>This is a decoy page</html>"
	p2.UserId = 1
	models.PostPage(&p2)

	//add phishing page with proxy bypass enabled
	p3 := models.Page{Name: "Phishing Page"}
	p3.HTML = "<html>This is a phishing page</html>"
	p3.ProxyBypassEnabled = true
	p3.DecoyPageId = p2.Id
	p3.UserId = 1
	models.PostPage(&p3)

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

	// Setup and "launch" our campaign using proxy bypass
	// Set the status such that no emails are attempted
	c2 := models.Campaign{Name: "Proxy Bypass Test campaign"}
	c2.UserId = 1
	c2.Template = template
	c2.Page = p3
	c2.SMTP = smtp
	c2.Anchor = "11111111112222222222333333333312"
	c2.Groups = []models.Group{group}
	models.PostCampaign(&c2, c2.UserId)
	c2.UpdateStatus(models.CampaignEmailsSent)
}
