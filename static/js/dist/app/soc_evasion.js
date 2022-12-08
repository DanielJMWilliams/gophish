document.addEventListener("DOMContentLoaded", function(){
    console.log("loaded js");
    //get anchor and set it to be key
    var anchor = window.location.hash.substring(1);
    console.log("anchor: ", anchor)
    console.log(encrypted);
    //send decryption request to server
    // "thisis32bitlongpassphraseimusing"
    if(anchor!="" && encrypted!=""){
        data = {Text: encrypted, Key: anchor}
        $.ajax({
            url: 'http://127.0.0.1:80/api/decrypt',
            type: "POST",
            dataType: "json",
            crossDomain: true,
            data: JSON.stringify(data),
            success: function (data) {
                console.log(data.data);
                decryptedMessage = data.data["text"]
                console.log(decryptedMessage)
                //checks if it dcrypted successfully
                //WARNING: This may open up the encryption for analysis
                // could check if first char is < which opens up less info for analysis
                if(decryptedMessage.substring(0, 1)=="<"){
                    //sets the html to the decrypted message
                    document.write(decryptedMessage);            
                }
        

            },
            fail: function () {
                console.log("Encountered an error")
            }
        });
    }

});
