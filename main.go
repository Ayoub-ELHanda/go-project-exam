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

const workerCount = 1

var port = 1208

type UserRequest struct {
	User   string      `json:"User"`
	Secret interface{} `json:"Secret,omitempty"`
}

func checkPort(portChan chan int, wg *sync.WaitGroup, client *http.Client, user string) string {
	defer wg.Done()

	var userSecretResponse string // Variable to store the /getUserSecret response

	for port := range portChan {
		address := "10.49.122.144:" + strconv.Itoa(port)
		conn, err := net.DialTimeout("tcp", address, time.Millisecond*100)

		if err != nil {
			continue
		}
		conn.Close()

		baseURL := fmt.Sprintf("http://10.49.122.144:%d", port)

		doPost := func(url string, body []byte) string {
			resp, err := client.Post(url, "application/json", bytes.NewBuffer(body))
			if err != nil {
				fmt.Println("Post error:", err)
				return ""
			}
			defer resp.Body.Close()
			respBody, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("Response from %s: %s\n", url, string(respBody))
			return string(respBody)
		}

		signUpURL := baseURL + "/signup"
		signUpBody, _ := json.Marshal(UserRequest{User: user})
		doPost(signUpURL, signUpBody)

		checkURL := baseURL + "/check"
		checkBody, _ := json.Marshal(UserRequest{User: user})
		doPost(checkURL, checkBody)

		getUserSecretURL := baseURL + "/getUserSecret"
		getUserSecretBody, _ := json.Marshal(UserRequest{User: user})
		userSecretResponse = doPost(getUserSecretURL, getUserSecretBody)

		if userSecretResponse != "Really don't feel like working today huh..." {
			fmt.Printf("User secret: %s\n", userSecretResponse)
		}

		getUserLevelURL := baseURL + "/getUserLevel"
		getUserLevelBody, _ := json.Marshal(UserRequest{User: user, Secret: userSecretResponse})
		doPost(getUserLevelURL, getUserLevelBody)

		getUserPointsURL := baseURL + "/getUserPoints"
		getUserPointsBody, _ := json.Marshal(UserRequest{User: user})
		doPost(getUserPointsURL, getUserPointsBody)

		hintURL := baseURL + "/iNeedAHint"
		hintBody, _ := json.Marshal(UserRequest{User: user})
		doPost(hintURL, hintBody)
	}

	return userSecretResponse
}

func main() {
	client := &http.Client{
		Timeout: time.Second * 2,
	}

	user := "Dragon" // Initialize the user outside the loop

	for {
		var wg sync.WaitGroup
		portChan := make(chan int)

		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go checkPort(portChan, &wg, client, user)
		}

		portChan <- port
		close(portChan)
		wg.Wait()
	}
}
