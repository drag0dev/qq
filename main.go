package main

import (
    _ "fmt"
    "log"
    "github.com/joho/godotenv"
)

func init(){
    // load env
    _ = godotenv.Load(".env")

}

func main(){
    log.Print("qq")
}
