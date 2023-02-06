# 08/12/2022 - SOC Detection Evasion Integrated
* SOC Evasion Detection now works

TODO:
* Refactor code so EncryptGCM function not repeated
* Automatically create decoy page and make it nice

# 08/12/2022 - soc_evasion.js working
* javascript decrypts the value, and sets the page to the decrypted HTML

TODO:
* fix encrypt/decrypt APIs so they can encrypt more than one block (16 bytes)

# 07/12/2022 - Encryptiono APIs used by javascript
* Changed the crypto APIs to run on phishing server
* example of communication with api in soc_evasion.js

# 02/12/2022 - Encryption/Decryption APIs working
* /api/encrypt and /api/decrypt APIs now working
* unit tests for both endpoints

# 01/12/2022 - Almost working encryption API
* Made /api/encrypt endpoint in crypto.go
* decrypt endpoint will be needed for the javascript SOC Evasion script to decrypt the encrypted landing pages


# 30/11/2022 - Adding script tags to html 
* Code now adds SOC Evasion script tags to HTML upon creation of new landing page
## TODO:
* Update on PUT: add/remove script
* Update soc_evasion.js with actual decryption code
* encrypt all of the html with a given key
