# qq (quick question)
cli tool that grabs answers from stackoverflow for a specified question, mostly aimed at a simpler questions which even the docs say:
>For more general searching, you should use a proper internet search engine restricted to the domain of the site in question.
## building
    go build -o qq main.go
after building it you can put it into bin or $PATH or any other place way you want

## usage
~~~
qq "your question" "tag"
~~~
search appears to just match your question as a substring to title/body, that is why question should be kept simple and less words usually gets better results<br/>
tag is a language, framework, etc.<br/><br/>
<img src='https://s7.gifyu.com/images/f.gif'></img>

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