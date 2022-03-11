package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"encoding/json"
	"net/http"
)

//TODO: move bigger vars from stack to heap

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

var base_url string = "https://api.stackexchange.com/2.3/"
var key string = "iSj2NwARiXutCdbZY9vkKQ(("
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

func removeHTMLTags(str *string){
        // all text in <code> paint orange
        *str = strings.ReplaceAll(*str, "<code>", "\033[0;33m")
        *str = strings.ReplaceAll(*str, "</code>", "\033[0m")

        // remote html tags
        htmlRegex := regexp.MustCompile(`<(“[^”]*”|'[^’]*’|[^'”>])*>`)
        *str = string(htmlRegex.ReplaceAllString(*str, ""))

        // TODO: turn multiple empty lines into one
        // TODO: html entities to actual chars
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

    // clean up html tags
    for index := 0; index < len(resJSON.Items); index++{
        removeHTMLTags(&resJSON.Items[index].Body)
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
        fmt.Println("Enter e to exit qq")
        fmt.Printf("Enter the number of the answer you want to see (1-%d): ", noOfAnswers)
        var userInput string
        fmt.Scanln(&userInput)
        if strings.ToUpper(userInput) == "E"{
            os.Exit(0)
        }

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
    for index := 0; index < len(resJSON.Items);index++{
        removeHTMLTags(&resJSON.Items[index].Body)
    }

    return resJSON
}

func printBody(thread *searchItem){
    clearScreen()
    fmt.Printf("Title: %s\n", thread.Title)
    fmt.Printf("Link: %s\n", thread.Link)
    fmt.Print("-------------------------------------------------------\n")
    fmt.Print(thread.Body)
    fmt.Print("-------------------------------------------------------\n")
    fmt.Print("Press enter to conitnue> ")
    var temp string
    fmt.Scanln(&temp)
}

func displayDetailedThread(thread searchItem, answers threadInfo){
    printBody(&thread)
    fmt.Printf("Title: %-100s\n", thread.Title)
    fmt.Printf("URL to the thread: %s\n", thread.Link)
    var userInput []byte = make([]byte, 1)
    var index int
    var length int = len(answers.Items)

    // disable input buffering and do not display entered char on clearScreen
    // TODO: this most likely doesnt work on windows
    exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
    exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

    for{
        clearScreen()
        fmt.Printf("Answer %d.\n", index+1)
        fmt.Printf("Link: %s\n", answers.Items[index].Link)
        fmt.Print("-------------------------------------------------------\n")
        fmt.Print(answers.Items[index].Body)
        fmt.Print("-------------------------------------------------------\n")

        for{
            // TODO: c - comments
            // commands
            // n - next answer
            // p - previous answer
            // e - exit qq
            // b - new answer
            // d - question body
            // h - help
            fmt.Print("\r> ")
            os.Stdin.Read(userInput)

            // if enter is pressed
            if strings.ToUpper(string(userInput[0])) == "N" && (index + 1 < length){
                index++
                break
            }else if strings.ToUpper(string(userInput[0])) == "P" && (index>0){
                index--
                break
            }else if strings.ToUpper(string(userInput[0])) == "E"{
                exec.Command("stty", "-F", "/dev/tty", "echo").Run()
                os.Exit(0)
            }else if strings.ToUpper(string(userInput[0])) == "B"{
                exec.Command("stty", "-F", "/dev/tty", "echo").Run()
                return
            }else if strings.ToUpper(string(userInput[0])) == "D"{
                printBody(&thread)
            }else if strings.ToUpper(string(userInput[0])) == "H"{
                fmt.Println("\nn - next answer\np - previous answer\ne - exit qq\nb - new answer\nd - question body\nq - new question")
            }
        }
    }
}

func init(){
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

    for{
        // display and get desired answer
        var num int
        num = displayRes(threads)

        // get detailed thread
        answers := getDetailedThread(threads[num].Question_id)

        // display detailed thread
        displayDetailedThread(threads[num], answers)
    }
}
