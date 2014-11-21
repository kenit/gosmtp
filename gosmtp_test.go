package gosmtp

import (
	"testing"
	"fmt"
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
	c := sender.AddQueue(&Task{Subject:"This is a test mail",To:[]string{"kenit@surehigh.com.tw"},Content:[]byte(content)})
	
	for err := range c{
		fmt.Println(err)
	}
}