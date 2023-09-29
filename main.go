package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const workerCount = 1

var port = 5174

type UserRequest struct {
	User   string      `json:"User"`
	Secret interface{} `json:"Secret,omitempty"`
}

func checkPort(portChan chan int, wg *sync.WaitGroup, client *http.Client, user string) {
	defer wg.Done()

	var secret string // Variable to store the user secret

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

		for i := 0; i < 15; i++ {
			getUserSecretURL := baseURL + "/getUserSecret"
			getUserSecretBody, _ := json.Marshal(UserRequest{User: user})
			respString := doPost(getUserSecretURL, getUserSecretBody)
			fmt.Printf("Iteration %d, Response from %s: %s\n", i, getUserSecretURL, respString) // Displaying the response

			if len(respString) > 45 {
				secret = strings.TrimSpace(strings.Split(respString, ": ")[1]) // Extracting the secret
				break                                                          // Exiting the loop once the secret is obtained
			}
		}

		if secret != "" {
			getUserLevelURL := baseURL + "/getUserLevel"
			getUserLevelBody, _ := json.Marshal(UserRequest{User: user, Secret: secret})
			doPost(getUserLevelURL, getUserLevelBody)

			getUserPointsURL := baseURL + "/getUserPoints"
			getUserPointsBody, _ := json.Marshal(UserRequest{User: user, Secret: secret})
			doPost(getUserPointsURL, getUserPointsBody)

			getHintURL := baseURL + "/iNeedAHint"
			getHintBody, _ := json.Marshal(UserRequest{User: user, Secret: secret})
			doPost(getHintURL, getHintBody)
		}
	}
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
