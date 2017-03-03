package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"golang.org/x/net/html"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

//Make a generic parseBlock function where you pass *html.Tokenizer and a JSON with HTML block structure, tokenizer is passed as reference

func getJSON(url string) {
	resp, err := myClient.Get(url)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer resp.Body.Close()
	doc := html.NewTokenizer(resp.Body)
	recordliquestion := false //Entry Question Block
	recordVote := false       //Voting block
	recordRating := false     //Rating block
	recordSpanEntry := false  //Question block
	recordahref := false      //Link to question
	readquestion := false     //Question text
	recordCode := false       //Code snippet
	for tokenType := doc.Next(); tokenType != html.ErrorToken; tokenType = doc.Next() {
		switch {
		case tokenType == html.StartTagToken:
			t := doc.Token()

			if t.Data == "li" {
				for _, a := range t.Attr {
					if a.Key == "class" && a.Val == "question" {
						recordliquestion = true
						break
					}
				}
				break
			}

			if t.Data == "div" && recordliquestion == true {
				for _, a := range t.Attr {
					if a.Key == "class" && a.Val == "votesNetQuestion" {
						recordVote = true
						break
					}
				}
				break
			}

			if t.Data == "span" && recordliquestion == true {
				for _, a := range t.Attr {
					if a.Key == "class" && a.Val == "entry" {
						recordSpanEntry = true
						break
					}
					if a.Key == "class" && a.Val == "commentCount" {
						recordRating = true
						break
					}

				}
				break
			}

			if t.Data == "a" && recordSpanEntry == true {
				for _, a := range t.Attr {
					if a.Key == "href" && strings.HasPrefix(a.Val, "/question?id=") {
						recordahref = true
						fmt.Println("URL: " + strings.Split(url, ".com")[0] + ".com" + a.Val)
						fmt.Println("ID: " + strings.TrimLeft(a.Val, "/question?id="))
						fmt.Println("Question:")
						break
					}
				}
				break
			}

			if t.Data == "p" && recordSpanEntry == true && recordahref == true {
				readquestion = true
				break
			}

			if t.Data == "code" && recordSpanEntry == true {
				recordCode = true
				break
			}

			if t.Data == "br" && recordSpanEntry == true && recordahref == true && readquestion == true {
				break
			}

		case tokenType == html.TextToken:
			t := doc.Token()
			if readquestion == true || recordCode == true {
				fmt.Print(t.Data)
				break
			}
			if recordVote == true {
				fmt.Println("Votes: " + t.Data)
				recordVote = false
				break
			}
			if recordRating == true {
				fmt.Println("Comments: " + t.Data)
				recordRating = false
				break
			}

		case tokenType == html.EndTagToken:
			t := doc.Token()
			if t.Data == "p" && readquestion == true {
				if recordCode == false {
					fmt.Println("")
				}
				readquestion = false
				recordahref = false
			}
			if t.Data == "code" && recordSpanEntry == true {
				recordCode = false
				break
			}

			if t.Data == "li" && recordliquestion == true {
				fmt.Println("\n*************************************")
				recordSpanEntry = false
				recordliquestion = false
			}
		}
	}
}

func main() {
	getJSON("https://www.careercup.com/page?pid=facebook-interview-questions&n=2")
}
