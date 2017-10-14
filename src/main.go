package main

import (
	"./message"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
	"errors"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) // seed rand so we get different values each time	
	msg := message.New("Guess a number between 1 and 20", "", "You haven't guessed yet!")
	port := getPort()

	
	// I consulted this article http://jessekallhoff.com/2013/04/14/go-web-apps-serving-static-files/
	// on how to serve specifc html pages. Not just index in the specified folder.
	http.HandleFunc("/guess/", func(w http.ResponseWriter, r *http.Request) {

		if !hasCookies(r) {
			// generate a new number between 1 - 20
			// I consulted this article https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/06.1.html
			// to learn how to set cookies in the response
			randomNum := rand.Intn(20) + 1                     // number between 1 - 20
			expiration := time.Now().Add(365 * 24 * time.Hour) // expires in 1 year
			cookie := http.Cookie{Name: "target", Value: strconv.Itoa(randomNum), Expires: expiration}
			http.SetCookie(w, &cookie)
		} // otherwise leave "target" at the current value

		var usersGuess string
		if userHasGuess(r) {
			usersGuess = r.URL.Query().Get("guess")
			msg.Guess = string(usersGuess)
		}

		targetCookie, err := getTargetCookie(r)
		if err == nil { // cookie exists
			targetAsInt, _ := strconv.Atoi(targetCookie.Value)
			userGuessAsInt, _ := strconv.Atoi(usersGuess)

			if userGuessAsInt < targetAsInt {
				msg.DisplayMessage = "You need to guess higher!"
			} else if userGuessAsInt > targetAsInt {
				msg.DisplayMessage = "You need to guess lower!"
			} else {
				msg.DisplayMessage = "You guessed the number correctly!"
				target, _ := getTargetCookie(r)                       // update existing cookie instead of adding one with duplicate name
				target.Value = strconv.Itoa(rand.Intn(20) + 1)        // generate a new random number
				target.Expires = time.Now().Add(365 * 24 * time.Hour) // reset expiry date
				http.SetCookie(w, target) // save the cookie so this new target is saved

			}
		}
		
		// create the template from the template file.
		guessTemplate := template.Must( // program will throw an error if there is a problem here
			template.ParseFiles("./html/guess.tmpl"))
		guessTemplate.Execute(w, msg)
	})
	// serves index.html in the res folder.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./html/index.html")
	})

	http.ListenAndServe(":"+port, nil)
}

func getTargetCookie(r *http.Request) (*http.Cookie, error) {
	cookies := r.Cookies() // all the cookies in the response
	for _, cookie := range cookies {
		if cookie.Name == "target" {
			return cookie, nil
		}
	}

	return nil, errors.New("Cookie does not exist.")
}

func userHasGuess(r *http.Request) bool {
	return len(r.URL.Query().Get("guess")) != 0
}

func hasCookies(r *http.Request) bool {
	return len(r.Cookies()) != 0
}

func getPort() string {
	// take in all the arguments bar the name of the file.
	args := os.Args[1:]
	var port string
	if len(args) == 1 { // user provided one command line argument
		if _, err := strconv.Atoi(args[0]); err == nil {
			port = args[0] // can be successfully parsed as a number.
		} else { // invalid command line argument provided.
			fmt.Println("Please provide a valid port number.\nExiting...")
			os.Exit(0)
		}
	} else {
		fmt.Println("Using default port value of 7777")
		port = "7777"
	}
	return port
}
