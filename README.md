# qq (quick question)
cli tool that grabs answers from stackoverflow for a specified question
## building
    go build -o qq main.go
after building it you can put it into bin or $PATH or any other place way you want

## usage
<img src='https://s7.gifyu.com/images/qq43abdb7067f01290.gif'></img>

## commands
**n** - next answer <br/>
**p** - previous answer <br/>
**e** - exit qq <br/>
**b** - new answer <br/>
**d** - question body <br/>
**c** - toggle between comments and answer <br />
**h** - help <br/>

## windows
as far as I could find there is no way to disable input buffering and display of characters being entered on windows, therefore it does not work on windows