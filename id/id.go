package id

import (
	"encoding/json"
	"fmt"
	"os"
)

// define identity for paypal api requests
type IdentConfig struct {
	ClientID     string `json:identID`
	ClientSecret string `json:identSecret`
}

func (ic IdentConfig) IdnetInfo() (string, string) {
	return ic.ClientID, ic.ClientSecret
}

func GetID() IdentConfig {
	f, err := os.Open(".id")
	if err != nil {
		panic(err)
	}
	var ic IdentConfig
	dec := json.NewDecoder(f)
	err = dec.Decode(&ic)
	if err != nil {
		panic(err)
	}
	fmt.Println("loaded .id")
	return ic
}
