package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
	"math/rand"
)






func handler(w http.ResponseWriter, r *http.Request) {
	responses := []string{`<?xml version="1.0" encoding="utf-8"?>
<soap12:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
  <soap12:Body>
    <createOrderResponse xmlns="http://dc.artistservices.com/crowdsurge/v1/">
      <createOrderResult>OK</createOrderResult>
    </createOrderResponse>
  </soap12:Body>
</soap12:Envelope>"`,
`<?xml version="1.0" encoding="utf-8"?>
<soap12:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
  <soap12:Body>
    <createOrderResponse xmlns="http://dc.artistservices.com/crowdsurge/v1/">
      <createOrderResult>BAD</createOrderResult>
    </createOrderResponse>
  </soap12:Body>
</soap12:Envelope>"`}
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
    if r.Method == "POST" {
        // receive posted data
        body, _ := ioutil.ReadAll(r.Body)
        // print for debugging purposes
        fmt.Println(string(body))
        i := rand.Int31n(2)
        fmt.Fprintf(w, "%s", responses[i])
        
        
        
    }
}



func main() {
    http.HandleFunc("/", handler)
    fmt.Println("running")
    http.ListenAndServe(":8080", nil)
}