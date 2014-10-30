package gosmtp

import (
	"testing"
	"fmt"
	"time"
)

func TestMail(t *testing.T) {
	sender := &EmailSender{ServerAddr:"172.16.1.10:25",SenderEmail:"kenit@surehigh.com.tw"}
	errChan := sender.Init()
	go func(){
		for err := range errChan{
			fmt.Println(err)
		}
	}()
	content:=`<html>
	<body>
	這是測試信的內容
	</body>
	</html>`
	c := sender.AddQueue(&Task{Subject:"這是一封測試信",To:[]string{"kenit@surehigh.com.tw"},Content:[]byte(content)})
	result := <- c
	fmt.Println(result)
	time.Sleep(20 * time.Second)
	content=`<html>
	<body>
	這是測試信的內容2
	</body>
	</html>`	
	c = sender.AddQueue(&Task{Subject:"這是一封測試信2",To:[]string{"kenit@surehigh.com.tw"},Content:[]byte(content)})
	result = <- c
	fmt.Println(result)	
}