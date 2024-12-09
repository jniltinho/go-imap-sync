package main


/*
Check for email from a certain sender (e.g john@coolcompany.net), grab the Excel attachment
Schedule this with cron or whatever...
*/

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)



func mainl() {
	ccc,err := client.DialTLS("imappro.zoho.com:993", nil)

	if err := ccc.Login("xxxxx", "xxxxx"); err != nil {
		fmt.Printf("Failed to login: %v", err)
	} else {
		fmt.Printf("Succesfully logged in...\n")
	}

	// Select INBOX
	mbox, err := ccc.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	// Get the last message
	if mbox.Messages == 0 {
		log.Fatal("No message in mailbox")
	} else {
		fmt.Printf("Total message(s) in inbox: %d\n", mbox.Messages)
	}
	seqSet := new(imap.SeqSet)
	seqSet.AddRange(1, mbox.Messages)

	// Get the whole message body
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 1)
	go func() {
		if err := ccc.Fetch(seqSet, items, messages); err != nil {
			log.Fatal(err)
		}
	}()

	for elem := range messages {
		if elem == nil {
			log.Fatal("Server didn't returned message")
		}

		r := elem.GetBody(&section)
		if r == nil {
			log.Fatal("Server didn't returned message body")
		}

		// Create a new mail reader
		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Fatal(err)
		}

		header := mr.Header

		if from, err := header.AddressList("From"); err == nil {

			if from[0].Address == "john@coolcompany.net" {
				if date, err := header.Date(); err == nil {

					theYear, theMonth, theDay := time.Now().Date()

					if theDay == date.Day() && theMonth == date.Month() && theYear == date.Year() {

						for {
							p, err := mr.NextPart()
							if err == io.EOF {
								break
							} else if err != nil {
								log.Fatal(err)
							}

							switch h := p.Header.(type) {
							case *mail.AttachmentHeader:
								filename, _ := h.Filename()
								// Okay we got the attachment
								savedAttachment, err := os.Create(filename)

								if err != nil {
									log.Fatal(err)
								}

								size, err := io.Copy(savedAttachment, p.Body)
								if err != nil {
									log.Fatal(err)
								}

								log.Printf("Saved %v bytes into %v\n", size, filename)
							}

						}
					}

				}
			}
		}
	}

	if err := ccc.Logout(); err != nil {
		fmt.Printf("Failed to logout: %v\n", err)
	}
}

