/*
	Function for launching a small web server that handles OAuth redirections from Raindrop.io,
	and thereby making OAuth authentication for a local Raindrop.io client possible.

	By Andreas Westerlind in 2021
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"
)

func authserver() {
	pid := fmt.Sprint(os.Getpid())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := ""
		querystring, _ := url.ParseQuery(r.URL.RawQuery)
		if len(querystring["code"]) > 0 {
			code = querystring["code"][0]
			if request_token(code) {
				http.ServeFile(w, r, "auth_info.html")
			} else {
				http.ServeFile(w, r, "auth_error.html")
			}
		} else {
			http.ServeFile(w, r, "auth_error.html")
		}
		endcmd := exec.Command("/bin/sh", "-c", "sleep 3 && pkill -P "+pid+" && kill "+pid, "&")
		endcmd.Start()
	})
	endcmd := exec.Command("/bin/sh", "-c", "sleep 1200 && kill "+pid, "&")
	endcmd.Start()
	http.ListenAndServe("127.0.0.1:11038", nil)
}

func request_token(code string) bool {
	// Prepare POST variables
	post_variables := url.Values{}
	post_variables.Add("code", code)
	post_variables.Add("client_id", client_id)
	post_variables.Add("client_secret", client_secret)
	post_variables.Add("redirect_uri", "http://127.0.0.1:11038")
	post_variables.Add("grant_type", "authorization_code")

	resp, err := http.PostForm("https://raindrop.io/oauth/access_token", post_variables)
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	token := RaindropToken{}
	json.Unmarshal([]byte(body), &token)

	if token.Error != "" {
		return false
	}

	location, _ := time.LoadLocation("UTC")
	token.CreationTime = time.Now().In(location).Format("2006-01-02 15:04:05")

	token_json, _ := json.Marshal(token)

	// Save to Keychain
	if err := wf.Keychain.Set("raindrop_token", string(token_json)); err != nil {
		return false
	}

	return true
}
