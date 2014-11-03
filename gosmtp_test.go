package gosmtp

import (
	"testing"
	"fmt"
	"time"
)

func TestMail(t *testing.T) {
	sender := &EmailSender{ServerAddr:"192.168.254.2",ServerPort:25,SenderEmail:"kenit@surehigh.com.tw"}
	errChan := sender.Init()
	go func(){
		for err := range errChan{
			fmt.Println(err)
		}
	}()
	content:=`<html>
	<body>
	This is a test mail.
	</body>
	</html>`
	sender.AddQueue(&Task{Subject:"This is a test mail",To:[]string{"kenit@surehigh.com.tw"},Content:[]byte(content)})
	fmt.Println("Waiting for 20 secs.")
	time.Sleep(20 * time.Second)
	content=`<html>
	<body>
	This is a test mail2.
	</body>
	</html>`	
	c := sender.AddQueue(&Task{Subject:"This is a test mail2",To:[]string{"kenit@surehigh.com.tw"},Content:[]byte(content)})
	
	for err := range c{
		fmt.Println(err)
	}
}