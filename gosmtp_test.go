package gosmtp

import (
	"testing"
	"fmt"
)

func TestMail(t *testing.T) {
        sender := &EmailSender{
                        ServerAddr:"smtp.mailgun.org",
                        ServerPort:587,
                        Username: "<username>",
                        Password: "<password>",
                        SenderEmail:"noreply@hotelnabe.com.tw",
                        UseTLS: false}
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
	task := &Task{
		Subject: "This is a Test Mail.",
		To: []string{"kenit@surehigh.com.tw"},
		Content: []byte(content),
		Headers: make(map[string]string)}

	task.Headers["X-Mailgun-Variables"] = `{"TEST": "TEST STRING"}`

	uuid, c := sender.AddQueue(task)

	fmt.Printf("Message %s is in queue.\n", uuid)

	for err := range c{
		fmt.Println(err)
	}
}