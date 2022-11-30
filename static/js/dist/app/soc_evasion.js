document.addEventListener("DOMContentLoaded", function(){
    console.log("loaded js");
    //get anchor and set it to be key
    var key = window.location.hash.substring(1);
    console.log("anchor: ", key)

});
