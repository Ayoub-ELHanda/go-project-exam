package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const workerCount = 100

type UserRequest struct {
	User   string      `json:"User"`
	Secret interface{} `json:"Secret,omitempty"` // Allow for different data types for the Secret
}

func checkPort(portChan chan int, wg *sync.WaitGroup, client *http.Client, user string, secret interface{}) {
	defer wg.Done()

	for port := range portChan {
		address := "10.49.122.144:" + strconv.Itoa(port)
		conn, err := net.DialTimeout("tcp", address, time.Millisecond*100)

		if err != nil {
			continue
		}
		conn.Close()

		baseURL := fmt.Sprintf("http://10.49.122.144:%d", port)

		doPost := func(url string, body []byte) {
			resp, err := client.Post(url, "application/json", bytes.NewBuffer(body))
			if err != nil {
				fmt.Println("Post error:", err)
				return
			}
			defer resp.Body.Close()
			respBody, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("Response from %s: %s\n", url, string(respBody))
		}

		signUpURL := baseURL + "/signup"
		signUpBody, _ := json.Marshal(UserRequest{User: user})
		doPost(signUpURL, signUpBody)

		checkURL := baseURL + "/check"
		checkBody, _ := json.Marshal(UserRequest{User: user})
		doPost(checkURL, checkBody)

		getUserSecretURL := baseURL + "/getUserSecret"
		getUserSecretBody, _ := json.Marshal(UserRequest{User: user})
		doPost(getUserSecretURL, getUserSecretBody)

		getUserLevelURL := baseURL + "/getUserLevel"
		getUserLevelBody, _ := json.Marshal(UserRequest{User: user})
		doPost(getUserLevelURL, getUserLevelBody)

		getUserPointsURL := baseURL + "/getUserPoints"
		getUserPointsBody, _ := json.Marshal(UserRequest{User: user})
		doPost(getUserPointsURL, getUserPointsBody)

		hintURL := baseURL + "/iNeedAHint"
		hintBody, _ := json.Marshal(UserRequest{User: user, Secret: secret}) // Pass the secret
		doPost(hintURL, hintBody)
	}
}

func main() {
	client := &http.Client{
		Timeout: time.Second * 2,
	}

	user := "ayoub"
	secret := "?" //

	for {
		var wg sync.WaitGroup
		portChan := make(chan int, workerCount)

		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go checkPort(portChan, &wg, client, user, secret)
		}

		for port := 1; port <= 65535; port++ {
			portChan <- port
		}

		close(portChan)
		wg.Wait()
	}
}
