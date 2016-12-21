package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type application struct {
	Job       string `json:"job"`
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

type attributes struct {
	Email string `json:"tas.personal.email"`
}

var fatalLog = log.New(os.Stdout, "FATAL: ", log.LstdFlags)
var infoLog = log.New(os.Stdout, "INFO: ", log.LstdFlags)

func basePage(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	email := getEmail(req.Header.Get("tazzy-tenant"), req.Header.Get("tazzy-saml"))
	t, err := template.ParseFiles("static/index.html")
	infoLog.Printf("BasePage template error: %v", err)
	t.Execute(rw, application{
		Job:   vars["job"],
		Email: email,
	})
}

func submit(rw http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		return
	}
	app := application{
		Job:       req.FormValue("Job"),
		Email:     req.FormValue("Email"),
		FirstName: req.FormValue("FirstName"),
		LastName:  req.FormValue("LastName"),
	}
	data, err := json.Marshal(&app)
	if err != nil {
		infoLog.Printf("Submit json error: %v", err)
		http.Error(rw, "Could not serialize input", http.StatusInternalServerError)
		return
	}
	postHTTP(req.Header.Get("tazzy-tenant"), getURL("devs/allan/submit"), data)
	http.Redirect(rw, req, getURL("devs/allan/board"), 301)
}

func getEmail(tenant, saml string) string {
	url := getURL(fmt.Sprintf("core/tenants/%s/saml/assertions/byKey/%s/json", tenant, saml))
	jsonAttr, err := getHTTP(tenant, url)
	infoLog.Printf("GetEmail json error: %v", err)
	if err != nil {
		return ""
	}

	var attr attributes
	infoLog.Printf("GetEmail attr error: %v", json.Unmarshal(jsonAttr, &attr))
	return attr.Email
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/tas/devs/allan/apply/{job}", basePage)
	r.HandleFunc("/submit", submit)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	fatalLog.Fatal(http.ListenAndServe(":8080", r))
}

func postHTTP(tenant, url string, data []byte) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	return doHTTP(req, tenant)
}

func getHTTP(tenant, url string) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Content-Type", "application/json")
	return doHTTP(req, tenant)
}

func doHTTP(req *http.Request, tenant string) ([]byte, error) {
	req.Header.Set("tazzy-secret", os.Getenv("IO_TAZZY_SECRET"))
	req.Header.Set("tazzy-tenant", tenant)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func getURL(api string) string {
	return fmt.Sprintf("%s/%s", os.Getenv("IO_TAZZY_URL"), api)
}
