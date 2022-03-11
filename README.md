# qq (quick question)
cli tool that grabs answers from stackoverflow for a specified question
## building
    go build -o qq main.go
after bulding it you can put into /usr/bin or $PATH or any other place way you want

## usage
    qq "how to center a div"
call would looks like this, if only one word is to be queried quotes are not required

## commands
**n** - next answer <br/>
**p** - previous answer <br/>
**e** - exit qq <br/>
**b** - new answer <br/>
**d** - question body <br/>
**h** - help <br/>

## windows
it may not work