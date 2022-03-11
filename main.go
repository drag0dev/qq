package main

import (
	"fmt"
	"io/ioutil"
	"os"
  "os/exec"
  "runtime"
	"strings"
  "strconv"
  "regexp"

	"encoding/json"
	_ "encoding/json"
	"net/http"

	"github.com/joho/godotenv"
)

type searchItem struct{
    Tags []string               `json:"tags"`
    Last_activity_date int64    `json:"last_activity_date"`
    Question_id int64           `json:"question_id"`
    Link string                 `json:"link"`
    Title string                `json:"title"`
    Body string                 `json:"body"`
}

type searchRes struct{
    Items []searchItem    `json:"Items"`
}

var base_url string
var key string
var clear map[string]func()

func getUserInput(question *string){
    var args []string = os.Args[1:]

    // for now only one arg and that is the question
    if len(args) != 1 {
        fmt.Println("Invalid arguments, exiting!")
        os.Exit(0)
    }

    *question = strings.ReplaceAll(args[0], " ", "%20")
}

func getSearchRes(question string)([]searchItem){
    client := http.Client{}

    var url string = base_url + fmt.Sprintf("search?key=%s&order=desc&sort=votes&intitle=%s&site=stackoverflow&filter=!LaSRLv)IiArQZm_BTFPx*I", key, question)

    // make a new request
    req, err := http.NewRequest("get", url, nil)
    if err != nil {
        fmt.Printf("Error making a new request: %s\n", err)
        os.Exit(0)
    }

    req.Header.Set("Host", "api.stackexchange.com")

    // send the new request
    res, err := client.Do(req)
    if err != nil{
        fmt.Printf("Error getting search results: %s\n", err)
        os.Exit(0)
    }

    // check if res is 200
    if res.Status != "200 OK"{
        fmt.Printf("There is a problem getting answers! (%s)\n", res.Status)
        os.Exit(0)
    }

    // parsing the res body
    body, err := ioutil.ReadAll(res.Body)
    if err != nil{
        fmt.Printf("Error parsing response body: %s\n", err)
        os.Exit(0)
    }

    res.Body.Close()

    // unmarshaling json
    var resJSON searchRes
    err = json.Unmarshal([]byte(body), &resJSON)
    if err != nil{
        fmt.Printf("Error parsing json: %s\n", err)
        os.Exit(0)
    }

    return resJSON.Items
}

func clearScreen(){
    value, ok := clear[runtime.GOOS]
    if ok {
        value()
    }else {
        fmt.Println("Unsupported os, exiting!")
        os.Exit(0)
    }

}

func displayRes(res []searchItem) (int){
    // display resutls
    clearScreen()
    fmt.Printf("%-3s\t%-50s\t %s\n", "no", "title", "tags")
    var noOfAnswers int = 0
    for index, answers := range res{
        if index == 20{
            noOfAnswers = index
            break
        }
        var title string= answers.Title
        if len(title) > 50{
            title = title[0:49]
        }
        var tags string = ""
        for i, tag := range answers.Tags{
            if i == 3 {
                break
            }
            tags = tags + " " + tag
        }

        fmt.Printf("%d.\t%-50s\t%s\n", index+1, title, tags)
    }

    // take user input
    var num int
    var err error
    fmt.Print("\n\n\n")
    for {
        fmt.Printf("Enter the number of the answer you want to see (1-%d): ", noOfAnswers)
        var userInput string
        fmt.Scanln(&userInput)
        num, err = strconv.Atoi(userInput)
        if err != nil {
            fmt.Println("You need to enter a number!\n")
            continue
        }else{
            if num < 1 || num > noOfAnswers{
                fmt.Printf("You need to enter a number in range (1-%d)!\n\n", noOfAnswers)
            }else{
                // no to index
                num--
                break
            }
        }
    }

    return num
}

type threadInfo struct {
    Items []struct{
        Link string                 `json:"link"`
        Body string                 `json:"body"`
    }                               `json:"items"`
}

func getDetailedThread(questionId int64)(threadInfo){
    client := http.Client{}

    // filter
    // everything default except:
    // answer - body, link
    var url string = base_url + fmt.Sprintf("questions/%d/answers?key=%s&order=desc&sort=votes&site=stackoverflow&filter=!4(lY7-qjlgWB7Z01e", questionId, key)

    req, err := http.NewRequest("get", url, nil)
    if err != nil {
        fmt.Println("Error making a new reqest for the whole thread, exiting!")
        os.Exit(0)
    }

    res, err := client.Do(req)
    if err != nil {
        fmt.Println("Error making a request to get the whole thread, exiting!")
        os.Exit(0)
    }else if res.Status != "200 OK"{
        fmt.Printf("Error getting the whole thread (%s), exiting!\n", res.Status)
    }

    body, err := ioutil.ReadAll(res.Body)
    if err != nil{
        fmt.Print("Error reading response body, exiting!")
        os.Exit(0)
    }
    res.Body.Close()

    var resJSON threadInfo
    err = json.Unmarshal([]byte(body), &resJSON)
    if err != nil{
        fmt.Println("Error parsing body of the response, exiting!")
        os.Exit(0)
    }

    // remove html elements
    // TODO: print text inside <code> different color
    for index, answer := range resJSON.Items{
        htmlRegex := regexp.MustCompile(`<(“[^”]*”|'[^’]*’|[^'”>])*>`)
        resJSON.Items[index].Body = string(htmlRegex.ReplaceAllString(answer.Body, ""))
    }

    return resJSON
}

func displayDetailedThread(thread searchItem, answers threadInfo){
    fmt.Printf("Title: %-100s\n", thread.Title)
    fmt.Printf("URL to the thread: %s\n", thread.Link)
    var userInput string
    var index int
    var length int = len(answers.Items)

    for{
        clearScreen()
        fmt.Printf("Answer %d.\n", index+1)
        fmt.Printf("Link: %s\n", answers.Items[index].Link)
        fmt.Print("-------------------------------------------------------\n")
        fmt.Print(answers.Items[index].Body)
        fmt.Print("-------------------------------------------------------\n")

        fmt.Scanln(&userInput)
        if index +1 < length{
            index++
        }else{
            index = 0
        }

    }
}

func init(){
    // load env
    _ = godotenv.Load(".env")
    base_url = os.Getenv("BASE_URL")
    key = os.Getenv("KEY")

    clear = make(map[string]func())
    clear["linux"] = func () {
        cmd := exec.Command("clear")
        cmd.Stdout = os.Stdout
        cmd.Run()
    }

    clear["windows"] = func () {
        cmd := exec.Command("cmd", "/c", "cls")
        cmd.Stdout = os.Stdout
        cmd.Run()
    }
}

func main(){
    // get question
    var question string;
    getUserInput(&question)

    // query question
    var threads []searchItem
    threads = getSearchRes(question)

    // display and get desired answer
    var num int
    num = displayRes(threads)

    // get detailed thread
    answers := getDetailedThread(threads[num].Question_id)

    // display detialed thread
    displayDetailedThread(threads[num], answers)
}
