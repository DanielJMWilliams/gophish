package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	log "github.com/gophish/gophish/logger"
)

// Page contains the fields used for a Page model
type Page struct {
	Id                 int64     `json:"id" gorm:"column:id; primary_key:yes"`
	UserId             int64     `json:"-" gorm:"column:user_id"`
	Name               string    `json:"name"`
	HTML               string    `json:"html" gorm:"column:html"`
	CaptureCredentials bool      `json:"capture_credentials" gorm:"column:capture_credentials"`
	CapturePasswords   bool      `json:"capture_passwords" gorm:"column:capture_passwords"`
	RedirectURL        string    `json:"redirect_url" gorm:"column:redirect_url"`
	AnchorEncryption   bool      `json:"anchor_encryption" gorm:"column:anchor_encryption"`
	ModifiedDate       time.Time `json:"modified_date"`
}

// ErrPageNameNotSpecified is thrown if the name of the landing page is blank.
var ErrPageNameNotSpecified = errors.New("Page Name not specified")

// parseHTML parses the page HTML on save to handle the
// capturing (or lack thereof!) of credentials and passwords
func (p *Page) parseHTML() error {
	d, err := goquery.NewDocumentFromReader(strings.NewReader(p.HTML))
	if err != nil {
		return err
	}
	forms := d.Find("form")
	forms.Each(func(i int, f *goquery.Selection) {
		// We always want the submitted events to be
		// sent to our server
		f.SetAttr("action", "")
		if p.CaptureCredentials {
			// If we don't want to capture passwords,
			// find all the password fields and remove the "name" attribute.
			if !p.CapturePasswords {
				inputs := f.Find("input")
				inputs.Each(func(j int, input *goquery.Selection) {
					if t, _ := input.Attr("type"); strings.EqualFold(t, "password") {
						input.RemoveAttr("name")
					}
				})
			} else {
				// If the user chooses to re-enable the capture passwords setting,
				// we need to re-add the name attribute
				inputs := f.Find("input")
				inputs.Each(func(j int, input *goquery.Selection) {
					if t, _ := input.Attr("type"); strings.EqualFold(t, "password") {
						input.SetAttr("name", "password")
					}
				})
			}
		} else {
			// Otherwise, remove the name from all
			// inputs.
			inputFields := f.Find("input")
			inputFields.Each(func(j int, input *goquery.Selection) {
				input.RemoveAttr("name")
			})
		}
	})
	p.HTML, err = d.Html()
	return err
}

// Validate ensures that a page contains the appropriate details
func (p *Page) Validate() error {
	if p.Name == "" {
		return ErrPageNameNotSpecified
	}
	// If the user specifies to capture passwords,
	// we automatically capture credentials
	if p.CapturePasswords && !p.CaptureCredentials {
		p.CaptureCredentials = true
	}
	if err := ValidateTemplate(p.HTML); err != nil {
		return err
	}
	if err := ValidateTemplate(p.RedirectURL); err != nil {
		return err
	}
	return p.parseHTML()
}

// GetPages returns the pages owned by the given user.
func GetPages(uid int64) ([]Page, error) {
	ps := []Page{}
	err := db.Where("user_id=?", uid).Find(&ps).Error
	if err != nil {
		log.Error(err)
		return ps, err
	}
	return ps, err
}

// GetPage returns the page, if it exists, specified by the given id and user_id.
func GetPage(id int64, uid int64) (Page, error) {
	log.Info("GETTING PAGE %d", id)
	p := Page{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&p).Error
	if err != nil {
		log.Error(err)
	}

	//embed html in innocent landing page if anchor encryption turned on
	if p.AnchorEncryption {
		p.HTML, err = EmbedEncryptedPage(p.HTML)
		log.Info(p.HTML)
	}

	return p, err
}

// GetPageByName returns the page, if it exists, specified by the given name and user_id.
func GetPageByName(n string, uid int64) (Page, error) {
	p := Page{}
	err := db.Where("user_id=? and name=?", uid, n).Find(&p).Error
	if err != nil {
		log.Error(err)
	}
	return p, err
}

// PostPage creates a new page in the database.
func PostPage(p *Page) error {
	log.Infof("Anchor encryption %t", p.HTML)
	err := p.Validate()
	if err != nil {
		log.Error(err)
		return err
	}
	// Inject anchor encryption script into HTML if option checked
	if p.AnchorEncryption {
		AddAnchorEncryptionScript(p)
	}
	// Insert into the DB
	err = db.Save(p).Error
	if err != nil {
		log.Error(err)
	}
	return err
}

func AddAnchorEncryptionScript(p *Page) {
	p.HTML = p.HTML + "<script src=\"https://127.0.0.1:3333/js/dist/app/soc_evasion.js\"></script>"
}
func RemoveAnchorEncryptionScript(p *Page) {
	scriptString := "<script src=\"https://127.0.0.1:3333/js/dist/app/soc_evasion.js\"></script>"
	p.HTML = strings.Replace(p.HTML, scriptString, "", 1)

}

// SOURCE: https://golangdocs.com/aes-encryption-decryption-in-golang
func Encrypt(html string) string {
	// cipher key
	key := []byte("thisis32bitlongpassphraseimusing")
	c := EncryptGCM(html, key)
	log.Info("ENCRYPTED: ")
	log.Info(c)
	return c
}

func EncryptGCM(stringToEncrypt string, keyString []byte) (encryptedString string) {
	log.Info("ENCRYPTGCM: ")
	log.Info(keyString)
	//Since the key is in string, we need to convert decode it to bytes
	key := keyString
	plaintext := []byte(stringToEncrypt)
	log.Info(key)
	log.Info(plaintext)
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

func EmbedEncryptedPage(html string) (string, error) {
	//encrypt all html and store in value in new html page
	// new html page will be innocent looking landing page
	encryptedHTML := Encrypt(html)

	// TODO: update parameters for all users and custom innocent page
	innocentPage, err := GetPageByName("InnocentPage", 1)

	if err != nil {
		return html, err
	}

	innocentPage.HTML += "<script>var encrypted = " + "\"" + encryptedHTML + "\"" + "</script>"

	return innocentPage.HTML, err

}

// PutPage edits an existing Page in the database.
// Per the PUT Method RFC, it presumes all data for a page is provided.
func PutPage(p *Page) error {
	err := p.Validate()
	if err != nil {
		return err
	}
	// Inject anchor encryption script into HTML if option checked
	if p.AnchorEncryption {
		AddAnchorEncryptionScript(p)
	} else {
		RemoveAnchorEncryptionScript(p)
	}

	err = db.Where("id=?", p.Id).Save(p).Error
	if err != nil {
		log.Error(err)
	}
	return err
}

// DeletePage deletes an existing page in the database.
// An error is returned if a page with the given user id and page id is not found.
func DeletePage(id int64, uid int64) error {
	err := db.Where("user_id=?", uid).Delete(Page{Id: id}).Error
	if err != nil {
		log.Error(err)
	}
	return err
}
