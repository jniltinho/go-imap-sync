package main

// go build -o list-folder clients/list-folder.go

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/jniltinho/gopass"
)

func getPassword(username, server string) (password string) {
	password = os.Getenv("IMAP_PASSWORD")

	if password == "" {
		log.Printf("Enter IMAP Password for %v on %v: ", username, server)
		passwordBytes, err := gopass.GetPasswd()
		if err != nil {
			panic(err)
		}
		password = string(passwordBytes)
	}
	return
}

func main() {

	var server, username string
	flag.StringVar(&server, "server", "", "sync from this mail server and port (e.g. mail.example.com:993)")
	flag.StringVar(&username, "username", "", "username for logging into the mail server")
	flag.Parse()

	if server == "" {
		log.Println("list-folder IMAP. Usage:")
		flag.PrintDefaults()
		log.Fatal("Required parameters not found.")
	}

	password := getPassword(username, server)

	parts := strings.Split(server, ":")
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Fatal(err)
	}

	// Connect to the server
	var c *client.Client
	if port == 143 {
		c, err = client.Dial(server)
	} else {
		c, err = client.DialTLS(server, nil)
	}

	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Login (replace username and password with your credentials)
	if err := c.Login(username, password); err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	for m := range mailboxes {
		fmt.Printf("- %s\n", m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	fmt.Println("Disconnected from IMAP server.")
}
