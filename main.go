package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	//	"os"
	"strings"
	//"flag"
	//	"time"
	"github.com/spf13/viper"
	"strconv"
)

type Domain struct {
	Subdomains []string `json:"subdomains"`
}

type AccountGroups struct {
	AccountGroups []struct {
		AccountGroupName string `json:"accountGroupName"`
		Aid              int    `json:"aid"`
		OrganizationName string `json:"organizationName"`
		Current          int    `json:"current"`
		Default          int    `json:"default"`
	} `json:"accountGroups"`
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

func GetAID(resp string) string {
	var teAccountGroups AccountGroups
	var aid = 0

	err := json.Unmarshal([]byte(resp), &teAccountGroups)
	if err != nil {
		fmt.Println(err)
	}
	for index, _ := range teAccountGroups.AccountGroups {
		if teAccountGroups.AccountGroups[index].Default == 1 {
			aid = teAccountGroups.AccountGroups[index].Aid
		}
	}
	if aid == 0 {
		panic(fmt.Errorf("Fatal error - AID wrong\n"))
	}

	return strconv.Itoa(aid)
}

func ValidateSubdomains (resp string) map[int]string {
	var domains Domain
	var teOauthToken = "1"
	ValidatedDomains := make(map[int]string)

	err := json.Unmarshal([]byte(resp), &domains)
	if err != nil {
		fmt.Println(err)
	}

	for index, _ := range domains.Subdomains {
		resp = GetRequest(viper.GetString("thousandeyes.validateUrl")+domains.Subdomains[index], teOauthToken)

		if strings.Contains(resp, "true"){
			ValidatedDomains[index] = domains.Subdomains[index]
		}		
	}

	Logger("Validated Sub-Domains: "+strconv.Itoa(len(ValidatedDomains)))
	return ValidatedDomains


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

	var teOauthToken = viper.GetString("thousandeyes.oauthToken")
	var teUser = viper.GetString("thousandeyes.user")
	var teDomain = viper.GetString("thousandeyes.domain")

	Logger("ThousandEyes Oauth Token: " + teOauthToken)
	Logger("ThousandEyes User: " + teUser)
	Logger("Domain of Interest: " + teDomain)

	Logger("Getting Sub-Domains")
	resp := GetRequest(viper.GetString("thousandeyes.serviceUrl")+teDomain, teOauthToken)
	
	//fmt.Println(resp)

	Logger("Validating Sub-Domains")
	ValidatedDomains := ValidateSubdomains(resp)
	fmt.Println(ValidatedDomains)

}
