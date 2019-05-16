package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

var (
	oauthConf = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "http://localhost:9090/home",
		Scopes:       []string{"public_profile"},
		Endpoint:     facebook.Endpoint,
	}
	oauthStateString = "thisshouldberandom"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Can not react environment configurations..")
		os.Exit(1)
	}
	oauthConf.ClientID = os.Getenv("facebook.clientID")
	oauthConf.ClientSecret = os.Getenv("facebook.clentSecret")
}

const htmlIndex = `<html><body>
Logged in with <a href="/login">facebook</a>
</body></html>
`

func handleMain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlIndex))
}

func handleFacebookLogin(w http.ResponseWriter, r *http.Request) {
	Url, err := url.Parse(oauthConf.Endpoint.AuthURL)
	if err != nil {
		log.Fatal("Parse: ", err)
	}
	fmt.Println(Url)
	parameters := url.Values{}
	parameters.Add("client_id", oauthConf.ClientID)
	parameters.Add("scope", strings.Join(oauthConf.Scopes, " "))
	parameters.Add("redirect_uri", oauthConf.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", oauthStateString)
	Url.RawQuery = parameters.Encode()
	url := Url.String()
	fmt.Println(url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleNavigateToHome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Home..")

	state := r.FormValue("state")
	fmt.Println(state)
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	fmt.Println(code)

	if code == "" {
		fmt.Println("Code not found..")
		w.Write([]byte("Code Not Found to provide AccessToken.."))
	} else {
		token, err := oauthConf.Exchange(oauth2.NoContext, code)
		if err != nil {
			fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
			return
		}
		fmt.Println("TOKEN>> AccessToken>>", token.AccessToken)
		fmt.Println("TOKEN>> Expiration Time>>", token.Expiry)
		fmt.Println("TOKEN>> RefreshToken>>", token.RefreshToken)

		resp, err := http.Get("https://graph.facebook.com/me?access_token=" +
			url.QueryEscape(token.AccessToken))
		if err != nil {
			fmt.Printf("Get: %s\n", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("ReadAll: %s\n", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		log.Printf("parseResponseBody: %s\n", string(response))

		w.Write([]byte("Hello, I'm protected\n"))
		w.Write([]byte(string(response)))
		return
	}
}

func main() {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", handleFacebookLogin)
	http.HandleFunc("/home", handleNavigateToHome)
	fmt.Print("Started running on http://localhost:9090\n")
	log.Fatal(http.ListenAndServe(":9090", nil))
}
