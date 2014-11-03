gosmtp
======

Usage:

sender := &EmailSender{ServerAddr:"192.168.254.2",ServerPort:25,SenderEmail:"sender@example.com"}
sender.Init()

content:=`<html>
	<body>
	This is a test mail.
	</body>
	</html>`
	
sender.AddQueue(&Task{Subject:"This is a test mail",To:[]string{"recipient@example.com"},Content:[]byte(content)})