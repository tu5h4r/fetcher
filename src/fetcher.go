
//============================================================================
// Name        : fetcher.gp
// Author      :  Tushar Singh
// Description : Async Non-DB solution for fetcher parser and web scrapper
//============================================================================

package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"golang.org/x/net/html"
	//"encoding/json"
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


type resultSet struct{
	ResultPageId int
	QuestionEntrySet []QuestionEntry
}

var myClient = &http.Client{Timeout: 4 * time.Second}

//Make a generic parseBlock function where you pass *html.Tokenizer and a JSON with HTML block structure, tokenizer is passed as reference

func getPageResult(url string,companyName string, pageID int, chFinished chan bool,  ret chan resultSet)  {

	link := "https://www.careercup.com/page?pid=" + companyName + "-interview-questions&n=" + strconv.Itoa(pageID)
	resp, err := myClient.Get(link)
	for err != nil {
		strError := err.Error()
		if( strings.Contains(strError,"Client.Timeout exceeded")){
			fmt.Println("Connection Timed Out - " + strError+"\nRetrying...")
			resp, err = myClient.Get(link)
		}else{
			log.Fatal(err)
			return //nil
		}
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
				if strings.Contains(t.Data,"<"){
					t.Data = strings.Replace(t.Data, "<", "< ", -1)
				}
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
	var x resultSet
	x.QuestionEntrySet = questionArray
	x.ResultPageId = pageID
	ret <- x
}

func runFetcher(company string, sortorder string) []CompanyQuestions {
	var cmpqs []CompanyQuestions
	var stqs []SiteQuestions
	
	var qsarr []resultSet
	qsarr = make([]resultSet,50)
	compq := make(chan resultSet)
	chFinished := make(chan bool)	

	for i := 1;i <= 50; i++ {  
		go getPageResult("https://www.careercup.com", company,i,chFinished,compq)		
	}

	for c := 1; c <= 50; {
		select {
		case ret := <-compq:
			qsarr[ret.ResultPageId-1] = ret
		case <-chFinished:
			c++
		}
	}
	for _, elem := range qsarr{
		stqs = append(stqs, SiteQuestions{"careercup",elem.QuestionEntrySet})
	}
	cmpqs = append(cmpqs,CompanyQuestions{company,stqs})
	return cmpqs
}


func ReceiveHttpRequest(w http.ResponseWriter, r *http.Request) {
	var htmlDATA string
	r.ParseForm()       // parse arguments, you have to call this by yourself
	fmt.Println(r.Form) // print form information in server side
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	
	htmlDATA = "<table><tr><td><select id =\"company\"><option value=amazon>Amazon</option>"
	htmlDATA += "<option value=google>Google</option>"
	htmlDATA += "<option value=facebook>Facebook</option>"
	htmlDATA += "<option value=linkedin>LinkedIn</option>"
	htmlDATA += "<option value=intel>Intel</option>"
	htmlDATA += "<option value=apple>Apple</option>"
	htmlDATA += "<option value=vmware-inc>VMWare Inc</option></select></td>"
	
	htmlDATA += "<td><select id =\"sorted\"><option value=date>Sort By Date</option>"
	htmlDATA += "<option value=comments>Sort By Comments</option>"
	htmlDATA += "<option value=votes>Sort By Votes</option></select></td>"
	htmlDATA += "<td><input type=\"submit\" value=\"Go\""
	htmlDATA += "onclick=\"location.href='?company='+getElementById('company').options[getElementById('company').selectedIndex].value"
	htmlDATA += "+'&sort='+getElementById('sorted').options[getElementById('sorted').selectedIndex].value"
	htmlDATA += "\"></td></tr></table>"
	

	if r.Form.Get("company") == "" || r.Form.Get("sort") == ""{
		fmt.Fprintln(w,htmlDATA)
		return
	}
	
	var st []CompanyQuestions 
	st = runFetcher(r.Form.Get("company"),r.Form.Get("sort"))
	
	htmlDATA += "<table border=\"1\" width=\"100%\" style= \"table-layout: fixed\">\n"
	htmlDATA += "<tr align=\"center\" verticalvalign=\"center\"><td><h3>"+strings.ToUpper(r.Form.Get("company"))+"</h3></td></tr>"
	for _,index := range st[0].SiteQuestionsSet{
		for _, elem := range index.QuestionEntrySet {
			//fmt.Println(elem.Date)
			htmlDATA += "<tr><td><table>\n"
			htmlDATA += "<tr><td><pre style=\"white-space: pre-wrap;word-wrap: break-word\">\n"+ elem.Question + "\n</pre>\n</td></tr>"
			htmlDATA += "<tr><td style=\"font-family:verdana;font-size:12 \">URL - <a href=\"" + elem.URL + "\" rel=\"noopener noreferrer\" target=\"_blank\">" + elem.URL +"</a></td></tr>\n"
			htmlDATA += "<tr><td style=\"font-family:verdana;font-size:12 \">Date - " + elem.Date + "</td></tr>\n"
			htmlDATA += "<tr><td style=\"font-family:verdana;font-size:12 \">Comments - " + elem.Comments + "</td></tr>\n"
			htmlDATA += "<tr><td style=\"font-family:verdana;font-size:12 \">Votes - " + elem.Votes + "</td></tr>\n"
			htmlDATA += "</table></td></tr>\n"
		}
	}
	htmlDATA += "</table>"	
	fmt.Fprintln(w,htmlDATA)
}

func main() {
	http.HandleFunc("/", ReceiveHttpRequest)       // set router
	err := http.ListenAndServe(":9090", nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
