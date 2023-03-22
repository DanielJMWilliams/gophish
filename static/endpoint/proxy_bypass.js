document.addEventListener("DOMContentLoaded", function(){
    //get anchor and set it to be key
    var anchor = window.location.hash.substring(1);
    //send decryption request to server
    if(anchor!="" && encrypted!=""){
        decrypt(encrypted, anchor).then((decryptedMessage) => {
            //checks if it decrypted successfully - all html pages will start with "<"
            if(decryptedMessage.substring(0, 1)=="<"){
                //sets the html to the decrypted message
                document.body.parentNode.setHTML(decryptedMessage)     
            }
            $('html').show()
        });
    }
    else{
        $('html').show()  
    }
    
});


const hexToBytes = (hex) => {
    const bytes = new Uint8Array(hex.length / 2);
    for (let i = 0; i < bytes.length; i++) {
      bytes[i] = parseInt(hex.substr(i * 2, 2), 16);
    }
    return bytes;
  };
  
  async function decrypt(ciphertext, key) {
    // Convert the ciphertext from a hex string to a Uint8Array
    ciphertext = hexToBytes(ciphertext);
    var enc = new TextEncoder();
    var encodedKey =  enc.encode(key);
    
    // Create a new Cipher Block from the key
    const block = await window.crypto.subtle.importKey(
      'raw',
      encodedKey,
      'AES-GCM',
      true,
      ['decrypt']
    );
  
    // Extract the nonce from the beginning of the ciphertext
    const nonce = ciphertext.slice(0, 12);
  
    // Extract the encrypted message from the ciphertext
    const encrypted = ciphertext.slice(12);
  
    // Decrypt the message using window.crypto.subtle.decrypt
    const decrypted = await window.crypto.subtle.decrypt(
      {
        name: 'AES-GCM',
        iv: nonce,
        tagLength: 128,
      },
      block,
      encrypted
    );
  
    // Decode the decrypted message to a string
    const decoder = new TextDecoder();
    const plaintext = decoder.decode(decrypted);
    return plaintext;
  }
  
