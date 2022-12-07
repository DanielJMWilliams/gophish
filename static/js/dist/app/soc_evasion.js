document.addEventListener("DOMContentLoaded", function(){
    console.log("loaded js");
    //get anchor and set it to be key
    var key = window.location.hash.substring(1);
    console.log("anchor: ", key)
    console.log(encrypted);
    //send decryption request to server
    data = {Text: "This is a secret", Key: "thisis32bitlongpassphraseimusing"}
    $.ajax({
        url: 'http://127.0.0.1:80/encrypt',
        type: "POST",
        dataType: "json",
        crossDomain: true,
        data: JSON.stringify(data),
        success: function (data) {
            console.log(data);
        },
        fail: function () {
            console.log("Encountered an error")
        }
      });


});
