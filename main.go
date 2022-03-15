package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
    "html"

	"encoding/json"
	"net/http"
)

type question struct{
    Tags []string               `json:"tags"`
    Comments []struct{
        Link string            `json:"link"`
        Body string             `json:"body"`
    }                           `json:"comments"`
    Last_activity_date int64    `json:"last_activity_date"`
    Question_id int64           `json:"question_id"`
    Link string                 `json:"link"`
    Title string                `json:"title"`
    Body string                 `json:"body"`
}

type searchRes struct{
    Items []question    `json:"Items"`
}

var base_url string = "https://api.stackexchange.com/2.3/"
var key string = "iSj2NwARiXutCdbZY9vkKQ(("
var clear map[string]func()

func getUserInput(question *string){
    var args []string = os.Args[1:]

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

        // remove multiple empty lines
        htmlRegex = regexp.MustCompile(`[\r\n]+`)
        *str = string(htmlRegex.ReplaceAllString(*str, "\n"))

        // remove newline at the beginning
        *str= strings.TrimPrefix(*str, "\n")

        // esacping html entities
        *str = html.UnescapeString(*str)
}

func getSearchRes(question string)([]question, []question){
    client := http.Client{}

    // search by title
    var url string = base_url + fmt.Sprintf("search?pagesize=20&key=%s&order=desc&sort=votes&intitle=%s&site=stackoverflow&filter=!tgYu)MVYQMRhXxIidh_Dm5kktzNkyDS", key, question)

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
    var resTitleJSON searchRes
    err = json.Unmarshal([]byte(body), &resTitleJSON)
    if err != nil{
        fmt.Printf("Error parsing json: %s\n", err)
        os.Exit(0)
    }

    // clean up html tags
    for i := 0; i < len(resTitleJSON.Items); i++{
        removeHTMLTags(&resTitleJSON.Items[i].Body)

        // clean up html in comment bodies
        for j := 0; j < len(resTitleJSON.Items[i].Comments); j++{
            removeHTMLTags(&resTitleJSON.Items[i].Comments[j].Body)
        }
    }

    // searching in body
    url = base_url + fmt.Sprintf("search?pagesize=20&key=%s&order=desc&sort=votes&intitle=%s&site=stackoverflow&filter=!tgYu)MVYQMRhXxIidh_Dm5kktzNkyDS", key, question)

    // make a new request
    req, err = http.NewRequest("get", url, nil)
    if err != nil{
        fmt.Println("Error making a request to get body search results, exiting!")
        os.Exit(0)
    }

    req.Header.Set("Host", "api.stackexchange.com")

    res, err = client.Do(req)
    if err != nil {
        fmt.Println("Error getting search by body results, exiting!")
        os.Exit(0)
    }else if res.Status != "200 OK"{
        fmt.Printf("Error getting search by body results, exiting! (%s)\n", res.Status)
        os.Exit(0)
    }

    body, err = ioutil.ReadAll(res.Body)
    if err != nil {
        fmt.Print("Error parsing response body, exiting!")
        os.Exit(0)
    }
    res.Body.Close()

    var resBodyJSON searchRes
    err = json.Unmarshal([]byte(body), &resBodyJSON)
    if err != nil {
        fmt.Printf("Error parsing json: %s\n", err)
        os.Exit(0)
    }

    // remove html from the body and the comments
    for i := 0; i < len(resBodyJSON.Items); i++{
        removeHTMLTags(&resBodyJSON.Items[i].Body)

        for j := 0; j < len(resBodyJSON.Items[i].Comments); j++{
            removeHTMLTags(&resBodyJSON.Items[i].Comments[j].Body)
        }
    }


    return resTitleJSON.Items, resBodyJSON.Items
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

func pickQuestion(titleRes []question, bodyRes []question) (int){
    // display resutls
    var userSelected int = 0
    var maxIndex int = len(titleRes) + len(bodyRes) - 1
    var userInput []byte = make([]byte, 1)

    // buffering and disable display
    exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
    exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

    for{
        clearScreen()

        // print title res
        fmt.Println("Results by title")
        fmt.Printf("%-3s\t%-50s\t %s\n", "no", "title", "tags")
        for index, answers := range titleRes{
            // if the current one is selected print it blue
            if index == userSelected{
                fmt.Print("\033[1;104m")
                fmt.Print("\033[1;91m")
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

            // closing escape char for coloring
            if index == userSelected{
                fmt.Print("\033[0m")
            }
        }

        // print body res
        fmt.Println("\nResults by body")
        fmt.Printf("%-3s\t%-50s\t %s\n", "no", "title", "tags")
        for index, answers := range bodyRes{
            // if the current one is selected print it blue
            if index+len(titleRes) == userSelected{
                fmt.Print("\033[1;104m")
                fmt.Print("\033[1;91m")
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

            fmt.Printf("%d.\t%-50s\t%s\n", len(titleRes) + index+1, title, tags)

            // closing escape char for coloring
            if index+len(titleRes) == userSelected{
                fmt.Print("\033[0m")
            }
        }

        fmt.Println()
        for{
            fmt.Print("\rUse j (go down) or k (go up) to select answer which you want and press enter, or type e to exit qq>")
            os.Stdin.Read(userInput)
            if strings.ToUpper(string(userInput[0])) == "J" && userSelected < maxIndex{
                userSelected++
                break
            }else if strings.ToUpper(string(userInput[0])) == "K" && userSelected > 0{
                userSelected--
                break
            }else if userInput[0]== 10 || userInput[0] == 13{
                exec.Command("stty", "-F", "/dev/tty", "echo").Run()
                clearScreen()
                fmt.Print("Grabbing answers...\n")
                return userSelected
            }else if strings.ToUpper(string(userInput[0])) == "E"{
                exec.Command("stty", "-F", "/dev/tty", "echo").Run()
                os.Exit(0)
            }
        }
    }
}
type questionAnswers struct {
    Items []struct{
        Answer_id int64             `json:"answer_id"`
        Link string                 `json:"link"`
        Body string                 `json:"body"`
    }                               `json:"items"`
}

type answerComment struct{
    Items []struct{
        Link string                 `json:"link"`
        Body string                 `json:"body"`
    }                               `json:"items"`
}

func getDetailedThread(questionId int64)(questionAnswers, map[int64]*answerComment){
    client := http.Client{}

    var url string = base_url + fmt.Sprintf("questions/%d/answers?key=%s&order=desc&sort=votes&site=stackoverflow&filter=!4(lY7-qjnWz1N.wT9", questionId, key)

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
        os.Exit(0)
    }

    body, err := ioutil.ReadAll(res.Body)
    if err != nil{
        fmt.Print("Error reading response body, exiting!")
        os.Exit(0)
    }
    res.Body.Close()

    var resJSON questionAnswers
    err = json.Unmarshal([]byte(body), &resJSON)
    if err != nil{
        fmt.Println("Error parsing body of the response, exiting!")
        os.Exit(0)
    }

    // remove html elements
    for index := 0; index < len(resJSON.Items);index++{
        removeHTMLTags(&resJSON.Items[index].Body)
    }

    // get answer comments
    comments := make(map[int64]*answerComment)
    for _, answer := range resJSON.Items{
        url = base_url + fmt.Sprintf(`answers/%d/comments?key=%s&order=desc&sort=votes&site=stackoverflow&filter=!bB.Oz07fFTfvm)`, answer.Answer_id, key)
        req, err := http.NewRequest("get", url, nil)
        if err != nil{
            fmt.Print("Error making a request to get comments, exiting!\n")
            os.Exit(0)
        }

        res, err := client.Do(req)
        if err != nil{
            fmt.Println("Error making a request to get comments, exiting!\n")
            os.Exit(0)
        }else if res.Status != "200 OK"{
            fmt.Printf("Error getting answer comments (%s), exiting!\n", res.Status)
            os.Exit(0)
        }

        body, err := ioutil.ReadAll(res.Body)
        if err != nil{
            fmt.Println("Error parsing body of the response, exiting!")
            os.Exit(0)
        }
        res.Body.Close()

        commentsJSON:= &answerComment{}
        err = json.Unmarshal([]byte(body), commentsJSON)
        if err != nil{
            fmt.Println("Error parsing body of the response, exiting!")
            os.Exit(0)
        }

        for _, comment := range commentsJSON.Items{
            // remove html from the comments
            removeHTMLTags(&comment.Body)
        }
        comments[answer.Answer_id] = commentsJSON
    }

    return resJSON, comments
}

func printBody(thread *question){
    clearScreen()
    fmt.Printf("Title: %s\n", thread.Title)
    fmt.Printf("Link: \033[0;34m %s \033[0m \n", thread.Link)
    fmt.Print("-------------------------------------------------------\n")
    fmt.Print(thread.Body)
    fmt.Print("-------------------------------------------------------\n")

    fmt.Println("\nComments:")
    for _, c := range thread.Comments{
        fmt.Print("-------------------------------------------------------\n")
        fmt.Printf("Link: \033[0;34m %s \033[0m \n", c.Link)
        fmt.Print("-------------------------------------------------------\n")
        fmt.Printf("Body: %s\n", c.Body)
        fmt.Print("-------------------------------------------------------\n\n")
    }

    var userInput []byte = make([]byte, 1)
    for{
        fmt.Print("\rPress anything to continue>")
        os.Stdin.Read(userInput)
        if userInput[0]!=0{
            break
        }
    }
}

func printComments(comments *answerComment){
    clearScreen()
    fmt.Println("Comments")
    for _, c := range comments.Items{
        fmt.Print("-------------------------------------------------------\n")
        fmt.Printf("Link: \033[0;34m %s \033[0m \n", c.Link)
        fmt.Print("-------------------------------------------------------\n")
        fmt.Printf("Body: %s\n", c.Body)
        fmt.Print("-------------------------------------------------------\n\n")
    }
    var userInput []byte = make([]byte, 1)
    for{
        fmt.Print("\rPress anything to continue>")
        os.Stdin.Read(userInput)
        if userInput[0]!=0{
            break
        }
    }
}

func displayDetailedThread(thread question, answers questionAnswers, comments map[int64]*answerComment){
    printBody(&thread)
    fmt.Printf("Title: %-100s\n", thread.Title)
    fmt.Printf("URL to the thread: %s\n", thread.Link)
    var userInput []byte = make([]byte, 1)
    var index int
    var length int = len(answers.Items)

    // disable input buffering and do not display entered char on clearScreen
    // there don't seem to be an easy way of doing the same for windows
    exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
    exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

    for{
        if len(answers.Items) > 0{
            clearScreen()
            fmt.Printf("Answer %d.\n", index+1)
            fmt.Printf("Link: %s\n", answers.Items[index].Link)
            fmt.Print("-------------------------------------------------------\n")
            fmt.Print(answers.Items[index].Body)
            fmt.Print("-------------------------------------------------------\n")
        }else{
            clearScreen()
            fmt.Println("There is no answers for this question!")
        }

        for{
            // commands
            // n - next answer
            // p - previous answer
            // e - exit qq
            // b - new answer
            // d - question body
            // c - toggle between comments and answer
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
                break
            }else if strings.ToUpper(string(userInput[0])) == "C"{
                printComments(comments[answers.Items[index].Answer_id])
                break
            }else if strings.ToUpper(string(userInput[0])) == "H"{
                fmt.Println("\nn - next answer\np - previous answer\ne - exit qq\nb - new answer\nd - question body\nc - toggle between answer and answer comments\nq - new question")
            }
        }
    } }

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

    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- c
        exec.Command("stty", "-F", "/dev/tty", "echo").Run()
        os.Exit(0)
    }()
}

func main(){
    // get question
    var questionStr string;
    getUserInput(&questionStr)

    // query question
    var titleRes, bodyRes []question
    titleRes, bodyRes = getSearchRes(questionStr)

    if len(titleRes) ==0 && len(bodyRes) ==0{
        clearScreen()
        fmt.Println("No answers could be found, exiting!")
        os.Exit(0)
    }

    for{
        // display and get desired answer
        var num int
        num = pickQuestion(titleRes, bodyRes)

        numTitleResults := len(titleRes)

        var answers questionAnswers
        var comments map[int64]*answerComment

        if num >= numTitleResults{
            num = num - numTitleResults
            answers, comments = getDetailedThread(bodyRes[num].Question_id)
            displayDetailedThread(bodyRes[num], answers, comments)
        }else {
            answers, comments = getDetailedThread(titleRes[num].Question_id)
            displayDetailedThread(titleRes[num], answers, comments)
        }
    }
}
