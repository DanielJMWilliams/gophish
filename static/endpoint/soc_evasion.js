document.addEventListener("DOMContentLoaded", function(){
    //get anchor and set it to be key
    var anchor = window.location.hash.substring(1);
    //send decryption request to server
    if(anchor!="" && encrypted!=""){
        data = {Text: encrypted, Key: anchor}
        $.ajax({
            url: 'http://127.0.0.1:80/api/decrypt',
            type: "POST",
            dataType: "json",
            crossDomain: true,
            data: JSON.stringify(data),
            success: function (data) {
                //console.log(data.data);
                decryptedMessage = data.data["text"]
                //checks if it decrypted successfully - all html pages will start with "<"
                if(decryptedMessage.substring(0, 1)=="<"){
                    //sets the html to the decrypted message
                    document.body.parentNode.setHTML(decryptedMessage)     
                }
                $('html').show()
            },
            fail: function () {
                console.log("Encountered a failure")
                $('html').show()  
            },
            error: function(){
                console.log("Encountered an error")
                $('html').show() 
            }
        });
    }
    else{
        $('html').show()  
    }
    
});
