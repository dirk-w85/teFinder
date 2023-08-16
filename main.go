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
	"github.com/spf13/viper"
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

func Logger(msg string) {
	if viper.GetBool("global.debug") {
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

func ValidateSubdomains (resp string) map[int]string {
	var domains Domain
	var teOauthToken = "1"
	ValidatedSubDomains := make(map[int]string)

	err := json.Unmarshal([]byte(resp), &domains)
	if err != nil {
		fmt.Println(err)
	}

	for index, _ := range domains.Subdomains {
		resp = GetRequest(viper.GetString("thousandeyes.validateUrl")+domains.Subdomains[index], teOauthToken)

		if strings.Contains(resp, "true"){
			ValidatedSubDomains[index] = domains.Subdomains[index]
		}		
	}

	Logger("Validated Sub-Domains: "+strconv.Itoa(len(ValidatedSubDomains)))
	return ValidatedSubDomains
}

func CreateTests(ValidatedSubDomains map[int]string, teOauthToken string, teAgentLabels string) {
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

	Logger("Label ID is: "+strconv.Itoa(labelID))

	resp = GetRequest("https://api.thousandeyes.com/v6/groups/"+strconv.Itoa(labelID)+".json", teOauthToken)

	err = json.Unmarshal([]byte(resp), &labelDetails)
	if err != nil {
		fmt.Println(err)
	}

	Logger("Label has "+strconv.Itoa(len(labelDetails.Groups[0].Agents))+" Agents")
	for index, _ := range labelDetails.Groups[0].Agents {
		agentIDs[index] = labelDetails.Groups[0].Agents[index].AgentID
	}

	Logger("Getting existing HTTP-Server Tests")
	resp = GetRequest("https://api.thousandeyes.com/v6/tests/http-server.json", teOauthToken)

	err = json.Unmarshal([]byte(resp), &existingTests)
	if err != nil {
		fmt.Println(err)
	}

	var testExists = false

	for index, _ := range ValidatedSubDomains {
		testExists = false
		Logger("Checking if Test exists: Servicefinder - https://"+ValidatedSubDomains[index])

		for index2, _ := range existingTests.Test {
			if existingTests.Test[index2].TestName == "Servicefinder - https://"+ValidatedSubDomains[index]{
				Logger("Tests exists already!")
				testExists = true
			}
		}

		if testExists == false {
			Logger("Tests does not exists - Creating!")
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

func main() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}

	if viper.GetBool("global.debug") {
		Logger("Debugging enabled")
	}

	//var teOauthToken = viper.GetString("thousandeyes.oauthToken")
	//var teUser = viper.GetString("thousandeyes.user")
	//var teDomain = viper.GetString("thousandeyes.domain")
	//var teAgentLabels = viper.GetStringMapString("thousandeyes.agentLabels")
	var teAgentLabels string = "Servicefinder"


	domaingPtr := flag.String("domain","none","Domain to be checked")
	teTokenPtr := flag.String("token","none","ThousandEyes oAuth Token")

	flag.Parse()

	teDomain := *domaingPtr
	teOauthToken := *teTokenPtr

	if teDomain == "none" {               
		panic(fmt.Errorf("No Domain specified"))
	}

	if teOauthToken == "none" {               
		panic(fmt.Errorf("No oAuth Token specified"))
	}

	Logger("ThousandEyes Oauth Token: " + teOauthToken)
	//Logger("ThousandEyes User: " + teUser)
	Logger("Domain of Interest: " + teDomain)

	Logger("Getting Sub-Domains")
	resp := GetRequest(viper.GetString("thousandeyes.serviceUrl")+teDomain, teOauthToken)

	Logger("Validating Sub-Domains")
	ValidatedSubDomains := ValidateSubdomains(resp)

	CreateTests(ValidatedSubDomains, teOauthToken, teAgentLabels)

}
