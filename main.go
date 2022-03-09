package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"encoding/json"
	_ "encoding/json"
	"net/http"

	"github.com/joho/godotenv"
)

type ownerStruct struct{
    Account_id int          `json:"account_id"`
    Reputation int          `json:"reputation"`
    User_id int64           `json:"user_id"`
    User_type string        `json:"user_type"`
    Profile_image string    `json:"profile_image"`
    Display_name string     `json:"display_name"`
    Link string             `json:"link"`
}

type searchItem struct{
    Tags []string               `json:"tags"`
    Owner ownerStruct           `json:"owner"`
    Is_answered bool            `json:"is_answered"`
    View_count int64            `json:"view_count"`
    Closed_date int64           `json:"closed_date"`
    Answer_count int            `json:"answer_count"`
    Score int                   `json:"score"`
    Last_activity_date int64    `json:"last_activity_date"`
    Creation_date int64         `json:"creation_date"`
    Question_id int64           `json:"question_id"`
    Link string                 `json:"link"`
    Closed_reason string        `json:"closed_reason"`
    Title string                `json:"title"`
}

type searchRes struct{
    Items []searchItem    `json:"Items"`
}

var base_url string
var key string

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

    var url string = base_url + fmt.Sprintf("search?key=%s&order=desc&sort=activity&intitle=%s&site=stackoverflow", key, question)

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

    // unmarshaling json
    var resJSON searchRes
    err = json.Unmarshal([]byte(body), &resJSON)
    if err != nil{
        fmt.Printf("Error parsing json: %s\n", err)
        os.Exit(0)
    }

    return resJSON.Items
}

func init(){
    // load env
    _ = godotenv.Load(".env")
    base_url = os.Getenv("BASE_URL")
    key = os.Getenv("KEY")
}

func main(){
    var question string;
    getUserInput(&question)
    getSearchRes(question)
}
