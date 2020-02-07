package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"net/smtp"
	"io/ioutil"
	"net/mail"
	"html/template"
	"bytes"
	"strings"
	"errors"
)

// webhook handler func
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// map all incoming messages and parse to KV dict
	webhookData := make(map[string]interface{})
	// read from socket and save as raw request body
	reqBody, _err := ioutil.ReadAll(r.Body)
		if _err != nil {
			log.Println("Error processing Body...")
			log.Fatal(_err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	// try to parse request body as a JSON object
	err := json.Unmarshal(reqBody, &webhookData)
	if err !=nil {
		// if it's not a JSON, create new JSON
		log.Println("Input message is not a JSON")
		webhookData = map[string]interface{}{
			"Payload" : string(reqBody)}
	}
	//fmt.Println(reflect.TypeOf(webhookData))
	err = sendMail(webhookData)
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Send respone: ", http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		log.Println("Send respone: ", http.StatusOK)
		return
	}
}

func sendMail(data map[string]interface{}) (error) {
	
	//common
	servername := "host:port"
    from := mail.Address{"Notifier's alias", "alert_mail@group"}
	to   := []string{"impassive@mail.ru"}
	subj := "Kibana Alert Manager"
	mime := "MIME-version: 1.0;\nContent-type: text/html; charset=\"UTF-8\";\n\n"
	
	//headers
    headers := make(map[string]string)
    headers["From"] = from.String()
    headers["To"] = strings.Join(to, ",")
    headers["Subject"] = subj

    //start mail body build
    message := ""
	for k,v := range headers {
    	message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	
	message += mime

    //parse static template file
	tmpl, _ := template.ParseFiles("/go/bin/tpl/static/mail.html")

	//render data cause go can't work with dynamic map interfaces inside templates...
	tail := ""
	for k,v := range data {
    	tail += fmt.Sprintf("<tr><td>%s - %s</td></tr>", k, v)
	}
	//save IO from writer to param
	var tpl_byte bytes.Buffer
	err := tmpl.Execute(&tpl_byte, template.HTML(tail))
	if err != nil {
		log.Fatal(err)
		return errors.New("Failed to build html template!")
	}
	message += tpl_byte.String()
	//end mail body build

    // Connect to the remote SMTP server and dial channel
	c, err := smtp.Dial(servername)
	if err != nil {
		log.Fatal(err)
		return errors.New("Failed to Dial with an smtp server!")
	}
	//start formatting mail
	//add from partipiciant
	if err = c.Mail(from.Address); err != nil {
		log.Fatal(err)
		return errors.New("Failed to parse FROM section")
	}
	//add to patipiciant
	for _, t := range to {
        if err = c.Rcpt(t); err != nil {
			return errors.New("Failed to add partipiciants!")
        }
    }
	//Open data channel
	w, err := c.Data()
	if err != nil {
		log.Fatal(err)
		return errors.New("Failed to open Data channel to send mail!")
	}
	//write stream to data channel
	_, err = w.Write([]byte(message))
	if err != nil {
		log.Fatal(err)
		return errors.New("Failed to stream data!")
	}
	//close stream 
	err = w.Close()
	if err != nil {
		log.Fatal(err)
		return errors.New("Failed to close Data stream!")
	}
	//close connection
	c.Quit()
	//TODO: add more info to logger
	log.Println("Mail sent")
	return nil
}

// main
func main() {
	log.Println("server started")
	// create webhook handler
	http.HandleFunc("/webhook", handleWebhook)
	// startup server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
