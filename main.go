package main

import (
	"fmt"
	"mypay/api"
	"mypay/id"
)

func main() {
	// get identity
	fmt.Println("Getting api id")
	myID := id.GetID()
	// get apiString
	myKey := api.ReturnAPIkey(myID.ClientID, myID.ClientSecret)
	fmt.Println("Getting api key")
	fmt.Println(myKey)
}
