package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"golang.org/x/net/html"
	"encoding/json"
	"strconv"
)

//Singular question entry
type QuestionEntry struct {
    ID  string
    Votes string
    Comments  string
	URL string
	Date string
    Question string
	//Company string
	//Site string
}

//Per site collection of questions for a given company
type SiteQuestions struct {
	Site string // site name
	QuestionEntrySet []QuestionEntry
}

//Per company collection of questions across different sites
type CompanyQuestions struct {
	Company string
	SiteQuestionsSet []SiteQuestions
}


var myClient = &http.Client{Timeout: 100 * time.Second}

//Make a generic parseBlock function where you pass *html.Tokenizer and a JSON with HTML block structure, tokenizer is passed as reference

func getJSON(url string,companyName string, pageID string, chFinished chan bool,  ret chan []QuestionEntry)  {

	link := "https://www.careercup.com/page?pid=" + companyName + "-interview-questions&n=" + pageID
	fmt.Println(link)
	resp, err := myClient.Get(link)
	if err != nil {
		log.Fatal(err)
		return //nil
	}

	defer func() {
		// Notify that we're done after this function
		chFinished <- true
	}()

	defer resp.Body.Close()
	doc := html.NewTokenizer(resp.Body)
	recordliquestion := false //Entry Question Block
	recordVote := false       //Voting block
	recordRating := false     //Rating block
	recordSpanEntry := false  //Question block
	recordahref := false      //Link to question
	readquestion := false     //Question text
	recordCode := false       //Code snippet
	var qEntry QuestionEntry
	var questionText string
	var questionArray []QuestionEntry

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
						qEntry.URL = strings.Split(url, ".com")[0] + ".com" + a.Val
						qEntry.ID = strings.TrimLeft(a.Val, "/question?id=")
						break
					}
				}
				break
			}

			if t.Data == "abbr" && recordSpanEntry == true {
				for _, a := range t.Attr {
					if a.Key == "title" {
						qEntry.Date = a.Val
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
				questionText += "\n"
				break
			}

		case tokenType == html.TextToken:
			t := doc.Token()
			if readquestion == true || recordCode == true {
				questionText += t.Data
				break
			}
			if recordVote == true {
				qEntry.Votes = t.Data
				recordVote = false
				break
			}
			if recordRating == true {
				qEntry.Comments = t.Data
				recordRating = false
				break
			}

		case tokenType == html.EndTagToken:
			t := doc.Token()
			if t.Data == "p" && readquestion == true {
				questionText += "\n"
				readquestion = false
			}
			if t.Data == "code" && recordSpanEntry == true {
				recordCode = false
				break
			}
			if t.Data == "li" && recordliquestion == true {
				qEntry.Question = questionText
				questionText = ""
				questionArray = append(questionArray, qEntry)
				recordSpanEntry = false
				recordliquestion = false
			}
			if t.Data == "a" && recordSpanEntry == true {
				recordahref = false
			}
		}
	}

	ret <- questionArray
}

func main() {
	var cmpqs []CompanyQuestions
	var stqs []SiteQuestions
	var qsarr []QuestionEntry
	compq := make(chan []QuestionEntry)
	chFinished := make(chan bool)	

	for i := 1;i <= 14; i++ {  
		go getJSON("https://www.careercup.com", "google",strconv.Itoa(i),chFinished,compq)		
	}

	for c := 1; c <= 14; {
		select {
		case ret := <-compq:
			qsarr = append(qsarr,ret...)
		case <-chFinished:
			c++
		}
	}
	
	stqs = append(stqs, SiteQuestions{"careercup",qsarr})
	cmpqs = append(cmpqs,CompanyQuestions{"google",stqs})
	b, _ := json.MarshalIndent(cmpqs,"","  ")		

	fmt.Println("\n" + string(b))
}
