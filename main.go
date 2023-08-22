package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"bytes"
	"strings"
	"flag"
	"strconv"
) 

type Domain struct {
	Subdomains []string `json:"subdomains"`
}

type Label struct {
	Groups []struct {
		Name    string `json:"name"`
		GroupID int    `json:"groupId"`
		Type    string `json:"type"`
		Builtin int    `json:"builtin"`
	} `json:"groups"`
}

	
type LabelDetails struct {
	Groups []struct {
		Name    string `json:"name"`
		GroupID int    `json:"groupId"`
		Type    string `json:"type"`
		Builtin int    `json:"builtin"`
		Agents  []struct {
			AgentID     int      `json:"agentId"`
			AgentName   string   `json:"agentName"`
			AgentType   string   `json:"agentType"`
			CountryID   string   `json:"countryId"`
			TargetOnly  int      `json:"targetOnly"`
			IPAddresses []string `json:"ipAddresses"`
			Location    string   `json:"location"`
			Ipv6Policy  string   `json:"ipv6Policy"`
		} `json:"agents"`
	} `json:"groups"`
}

type Tests struct {
	Test []struct {
		Enabled             int    `json:"enabled"`
		TestID              int    `json:"testId"`
		TestName            string `json:"testName"`
		Interval            int    `json:"interval"`
		URL                 string `json:"url"`
		ModifiedDate        string `json:"modifiedDate"`
		NetworkMeasurements int    `json:"networkMeasurements"`
		CreatedBy           string `json:"createdBy"`
		ModifiedBy          string `json:"modifiedBy"`
		CreatedDate         string `json:"createdDate"`
	} `json:"test"`
}

type NewHttpServerTest struct {
	TestName				string
	Agents 					[]struct {
		AgentID 			int `json:"agentId"`
	}
	Interval 				int
	Url						string
}

func Logger(msg string, debugEnabled bool) {
	if debugEnabled {
		log.Println(msg)
	}
}

func GetRequest(url string, teToken string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("Authorization", "Bearer "+teToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return string(body)
}

func PostRequest(url string, teToken string, newTestString string) {

	jsonData := []byte(newTestString)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+teToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
}

func ValidateSubdomains (resp string, validateUrl string, debugEnabled bool) map[int]string {
	var domains Domain
	var teOauthToken = "1"
	ValidatedSubDomains := make(map[int]string)

	err := json.Unmarshal([]byte(resp), &domains)
	if err != nil {
		fmt.Println(err)
	}

	Logger("Sub-Domains to validate: "+strconv.Itoa(len(domains.Subdomains)),debugEnabled)

	for index, _ := range domains.Subdomains {
		resp = GetRequest(validateUrl+domains.Subdomains[index], teOauthToken)

		if strings.Contains(resp, "true"){
			ValidatedSubDomains[index] = domains.Subdomains[index]
		}		
	}

	Logger("Validated Sub-Domains: "+strconv.Itoa(len(ValidatedSubDomains)),debugEnabled)
	return ValidatedSubDomains
}

func CreateTests(ValidatedSubDomains map[int]string, teOauthToken string, teAgentLabels string,debugEnabled bool) {
	var labels Label
	var labelDetails LabelDetails
	var labelID int = 0
	var agentIDs = make(map[int]int)
	var existingTests Tests
	//var newTest NewHttpServerTest

	resp := GetRequest("https://api.thousandeyes.com/v6/groups.json", teOauthToken)
	
	err := json.Unmarshal([]byte(resp), &labels)
	if err != nil {
		fmt.Println(err)
	}

	for index, _ := range labels.Groups {
		if strings.Contains(labels.Groups[index].Name, teAgentLabels){
			labelID = labels.Groups[index].GroupID
		}
	}

	Logger("Label ID is: "+strconv.Itoa(labelID),debugEnabled)

	resp = GetRequest("https://api.thousandeyes.com/v6/groups/"+strconv.Itoa(labelID)+".json", teOauthToken)

	err = json.Unmarshal([]byte(resp), &labelDetails)
	if err != nil {
		fmt.Println(err)
	}

	Logger("Label has "+strconv.Itoa(len(labelDetails.Groups[0].Agents))+" Agents",debugEnabled)
	for index, _ := range labelDetails.Groups[0].Agents {
		agentIDs[index] = labelDetails.Groups[0].Agents[index].AgentID
	}

	Logger("Getting existing HTTP-Server Tests",debugEnabled)
	resp = GetRequest("https://api.thousandeyes.com/v6/tests/http-server.json", teOauthToken)

	err = json.Unmarshal([]byte(resp), &existingTests)
	if err != nil {
		fmt.Println(err)
	}

	var testExists = false

	for index, _ := range ValidatedSubDomains {
		testExists = false
		Logger("Checking if Test exists: Servicefinder - https://"+ValidatedSubDomains[index],debugEnabled)

		for index2, _ := range existingTests.Test {
			if existingTests.Test[index2].TestName == "Servicefinder - https://"+ValidatedSubDomains[index]{
				Logger("Tests exists already!",debugEnabled)
				testExists = true
			}
		}

		if testExists == false {
			Logger("Tests does not exists - Creating!",debugEnabled)
			newTestString := `{"testName":"Servicefinder - https://`+ValidatedSubDomains[index]+`","agents":[`
			
			for index, _ := range labelDetails.Groups[0].Agents {
				newTestString = newTestString+`{"agentId":`+strconv.Itoa(labelDetails.Groups[0].Agents[index].AgentID)+`},`
			}
			newTestString = strings.TrimRight(newTestString, ",")

			newTestString = newTestString+`],"interval":120,"url":"https://`+ValidatedSubDomains[index]+`"}`

			//fmt.Println(newTestString)

			PostRequest("https://api.thousandeyes.com/v6/tests/http-server/new.json", teOauthToken, newTestString)
		}		
	}
}

//------------------------------------
// GLOBALS
var serviceUrl string = "https://servicefinder.thousandeyes.com/retrieve-subdomains?domain="
var validateUrl string = "https://servicefinder.thousandeyes.com/valid-subdomain?subdomain="

//------------------------------------

func main() {

	domaingPtr := flag.String("domain","none","Domain to be checked")
	teTokenPtr := flag.String("token","none","ThousandEyes oAuth Token")
	teAgentLabelPtr := flag.String("agentlabel","none","ThousandEyes Agent Label")
	debugPtr := flag.Bool("debug",false,"ThousandEyes oAuth Token")

	flag.Parse()

	teDomain := *domaingPtr
	teOauthToken := *teTokenPtr
	debugEnabled := *debugPtr


	var teAgentLabels string = *teAgentLabelPtr //"Servicefinder"


	if debugEnabled {
		Logger("Debugging enabled", debugEnabled)
	}

	if teDomain == "none" {               
		panic(fmt.Errorf("No Domain specified"))
	}

	if teAgentLabels == "none" {               
		panic(fmt.Errorf("No Agent Label specified"))
	}

	if teOauthToken == "none" {               
		panic(fmt.Errorf("No oAuth Token specified"))
	}

	Logger("ThousandEyes Oauth Token: " + teOauthToken,debugEnabled)
	//Logger("ThousandEyes User: " + teUser)
	Logger("Domain of Interest: " + teDomain,debugEnabled)

	Logger("Getting Sub-Domains",debugEnabled)
	resp := GetRequest(serviceUrl+teDomain, teOauthToken)

	Logger("Validating Sub-Domains",debugEnabled)
	ValidatedSubDomains := ValidateSubdomains(resp,validateUrl,debugEnabled)

	CreateTests(ValidatedSubDomains, teOauthToken, teAgentLabels,debugEnabled)

}
