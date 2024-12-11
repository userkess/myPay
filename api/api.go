package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
)

// define data used
type tokenResponse struct {
	Scope        string
	Access_token string
	Token_type   string
	App_id       string
	Expires_in   time.Duration
	Nonce        string
}
type apiToken struct {
	AccessToken string
	ExpireTime  time.Time
}

func ReturnAPIkey(id string, secret string) string {
	// get current time
	currentTime := time.Now()

	// define identity
	clientID := id
	clientSecret := secret

	// init live apiToken
	var live apiToken
	// read token and expire time from db

	// open db connection
	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// get data
	db.View(func(tx *bolt.Tx) error {
		getData := tx.Bucket([]byte("DB"))
		live.AccessToken = string(getData.Get([]byte("token")))
		timeString := string(getData.Get([]byte("expire")))
		live.ExpireTime, _ = time.Parse(time.RFC3339, timeString)
		return nil
	})

	// get new token from Paypal (if needed)
	if live.ExpireTime.Before(currentTime) || live.AccessToken == "" {
		fmt.Println("updating expired token.")
		live.AccessToken, live.ExpireTime, err = getNewApiToken(clientID, clientSecret)
		// update new access token to db
		db.Update(func(tx *bolt.Tx) error {
			getData := tx.Bucket([]byte("DB"))
			err = getData.Put([]byte("token"), []byte(live.AccessToken))
			timeString := live.ExpireTime.Format(time.RFC3339)
			err = getData.Put([]byte("expire"), []byte(timeString))
			return err
		})
	}
	// return current working token
	return (live.AccessToken)
	//fmt.Println(live.ExpireTime)
}

// get new api tokens from Paypal
func getNewApiToken(id string, secret string) (string, time.Time, error) {

	// get current time
	currentTime := time.Now()

	// define Paypal request session
	contentType := "application/x-www-form-urlencoded"
	url := "https://api-m.sandbox.paypal.com/v1/oauth2/token"
	tokenRequest := []byte(`grant_type=client_credentials`)

	// format paypal auth string
	client := id + ":" + secret
	clientAuthString := base64.StdEncoding.EncodeToString([]byte(client))

	//build request to Paypal
	clientRequest := &http.Client{}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(tokenRequest))
	if err != nil {
		return "", currentTime, err
	}
	request.Header.Add("Content-Type", contentType)
	request.Header.Add("Authorization", "Basic "+clientAuthString)

	// send request
	response, err := clientRequest.Do(request)
	if err != nil {
		return "", currentTime, err
	}
	defer response.Body.Close()

	// get response
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", currentTime, err
	}

	// return response
	var responseValues tokenResponse
	json.Unmarshal(body, &responseValues)
	expireTime := (currentTime.Add(time.Second * responseValues.Expires_in))
	return responseValues.Access_token, expireTime, nil
}

// verify/create a db
func initDB() (*bolt.DB, error) {
	// open db
	db, err := bolt.Open("token.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db: %v", err)
	}
	// create base db
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("DB"))
		if err != nil {
			return fmt.Errorf("could not create DB: %v", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not setup database: %v", err)
	}
	return db, nil
}
