package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	gomail "gopkg.in/mail.v2"
)

type TomlConfig struct {
	SMTPServer string
	Port       int
	Username   string
	Password   string
	To         []string
}

var config TomlConfig

type OhMyFormSubmission struct {
	Form         int       `json:"form"`
	Submission   int       `json:"submission"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
	Fields       []struct {
		Field        int         `json:"field"`
		Slug         string      `json:"slug"`
		DefaultValue interface{} `json:"default_value"`
		Content      struct {
			Value interface{} `json:"value"`
		} `json:"content"`
	} `json:"fields"`
}

type QuickData struct {
	Name       string
	Age        string
	Org        string
	Email      string
	Smartphone string
	Newsletter string
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	var webhookData OhMyFormSubmission
	err = json.Unmarshal(body, &webhookData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tempQuickData := QuickData{}

	for _, field := range webhookData.Fields {
		switch field.Slug {
		case "name":
			tempQuickData.Name = fmt.Sprintf("%v", field.Content.Value)
		case "newsletter":
			tempQuickData.Newsletter = fmt.Sprintf("%v", field.Content.Value)
		case "org":
			tempQuickData.Org = fmt.Sprintf("%v", field.Content.Value)
		case "email":
			tempQuickData.Email = fmt.Sprintf("%v", field.Content.Value)
		case "age":
			tempQuickData.Age = fmt.Sprintf("%v", field.Content.Value)
		case "smartphone":
			tempQuickData.Smartphone = fmt.Sprintf("%v", field.Content.Value)
		}
	}

	log.Printf("Ayo someone just submitted the form %d - Name(%s) Age(%s) Email(%s) Org(%s) Smartphone(%s) Newsletter(%s)", webhookData.Form, tempQuickData.Name, tempQuickData.Age, tempQuickData.Email, tempQuickData.Org, tempQuickData.Smartphone, tempQuickData.Newsletter)
	sendMail(tempQuickData, webhookData.Form)
}

func sendMail(qd QuickData, form_no int) {
	body := fmt.Sprintf("<h3>Submission:</h3><br>Name: %s<br>Age: %s<br>Email: %s<br>Org: %s<br>Smartphone: %s<br>Newsletter: %s<br><hr><br>Take a look at <a href='https://forms2.sflc.in/admin/forms/%s'>form</a> or it's <a href='https://forms2.sflc.in/admin/forms/%s/submissions'>submission</a>.", qd.Name, qd.Age, qd.Email, qd.Org, qd.Smartphone, qd.Newsletter, strconv.Itoa(form_no), strconv.Itoa(form_no))
	m := gomail.NewMessage()

	m.SetHeader("From", config.Username)
	m.SetHeader("To", config.To...)
	m.SetHeader("Subject", qd.Name+" just filled form "+strconv.Itoa(form_no))
	m.SetBody("text/html", body)

	d := gomail.NewDialer(config.SMTPServer, config.Port, config.Username, config.Password)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
		panic(err)
	}
	log.Printf("Sent mail successfully!")
}

func main() {
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Listening on http://localhost:8080/ohmyformshook")
	http.HandleFunc("/ohmyformshook", handleWebhook)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
