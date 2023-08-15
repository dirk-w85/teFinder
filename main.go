package main

import (
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"encoding/json"
//	"os"
//	"strings"
	//"flag"
//	"time" 
	"strconv"
	"github.com/spf13/viper"
)


type AccountGroups struct {
	AccountGroups []struct {
		AccountGroupName string `json:"accountGroupName"`
		Aid              int    `json:"aid"`
		OrganizationName string `json:"organizationName"`
		Current          int    `json:"current"`
		Default          int    `json:"default"`
	} `json:"accountGroups"`
}

type Agents struct {
	Agents []struct {
		AgentID               int      `json:"agentId"`
		AgentName             string   `json:"agentName"`
		AgentType             string   `json:"agentType"`
		CountryID             string   `json:"countryId"`
		Enabled               int      `json:"enabled"`
		KeepBrowserCache      int      `json:"keepBrowserCache"`
		VerifySslCertificates int      `json:"verifySslCertificates"`
		IPAddresses           []string `json:"ipAddresses"`
		InterfaceIPMapping    []struct {
			InterfaceName string   `json:"interfaceName"`
			IPAddresses   []string `json:"ipAddresses"`
		} `json:"interfaceIpMapping"`
		LastSeen          string   `json:"lastSeen"`
		Location          string   `json:"location"`
		Network           string   `json:"network"`
		Prefix            string   `json:"prefix"`
		PublicIPAddresses []string `json:"publicIpAddresses"`
		TargetForTests    string   `json:"targetForTests"`
		AgentState        string   `json:"agentState"`
		Ipv6Policy        string   `json:"ipv6Policy"`
		Hostname          string   `json:"hostname"`
		CreatedDate       string   `json:"createdDate"`
		ErrorDetails      []any    `json:"errorDetails"`
	} `json:"agents"`
}

type Agent struct {
	AgentID int `json:"agentId"`
	Enabled int `json:"enabled"`
}

type Clusters struct {
	Agents []struct {
		AgentID        int    `json:"agentId"`
		AgentName      string `json:"agentName"`
		AgentType      string `json:"agentType"`
		ClusterMembers []struct {
			ErrorDetails []struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			} `json:"errorDetails"`
			IPAddresses       []string `json:"ipAddresses"`
			LastSeen          string   `json:"lastSeen"`
			MemberID          int      `json:"memberId"`
			Name              string   `json:"name"`
			Network           string   `json:"network"`
			Prefix            string   `json:"prefix"`
			PublicIPAddresses []string `json:"publicIpAddresses"`
			TargetForTests    string   `json:"targetForTests"`
			AgentState        string   `json:"agentState"`
		} `json:"clusterMembers"`
		CountryID             string `json:"countryId"`
		Enabled               int    `json:"enabled"`
		KeepBrowserCache      int    `json:"keepBrowserCache"`
		VerifySslCertificates int    `json:"verifySslCertificates"`
		Location              string `json:"location"`
		Ipv6Policy            string `json:"ipv6Policy"`
		CreatedDate           string `json:"createdDate"`
	} `json:"agents"`
}

type CloudAgents struct {
	Agents []struct {
		AgentID     int      `json:"agentId"`
		AgentName   string   `json:"agentName"`
		AgentType   string   `json:"agentType"`
		CountryID   string   `json:"countryId"`
		IPAddresses []string `json:"ipAddresses"`
		Location    string   `json:"location"`
	} `json:"agents"`
}

func Logger(msg string){
	if viper.GetBool("global.debug"){
		log.Println(msg)
	}
}

func GetRequest (url string, teToken string) string {
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

func GetEnterpriseAgents(resp string) map[int]int {
	var teEnterpriseAgents Agents
	EnterpriseAgentList := make(map[int]int)

	err := json.Unmarshal([]byte(resp), &teEnterpriseAgents)
	if err != nil {
		fmt.Println(err)
	}
	for index, _ := range teEnterpriseAgents.Agents {	
		if teEnterpriseAgents.Agents[index].Enabled == 1 {
			EnterpriseAgentList[index] = teEnterpriseAgents.Agents[index].AgentID
		}
	}
	return EnterpriseAgentList
}

func GetEnterpriseAgentsCluster(resp string) map[int]int {
	var teEnterpriseAgentClusters Clusters
	EnterpriseAgentClusterList := make(map[int]int)

	err := json.Unmarshal([]byte(resp), &teEnterpriseAgentClusters)
	if err != nil {
		fmt.Println(err)
	}
	for index, _ := range teEnterpriseAgentClusters.Agents {	
		if teEnterpriseAgentClusters.Agents[index].Enabled == 1 {
			EnterpriseAgentClusterList[index] = teEnterpriseAgentClusters.Agents[index].AgentID
		}
	}
	return EnterpriseAgentClusterList
}

func GetCloudAgents(resp string) map[int]int {
	var teCloudAgents CloudAgents
	CloudAgentList := make(map[int]int)

	err := json.Unmarshal([]byte(resp), &teCloudAgents)
	if err != nil {
		fmt.Println(err)
	}

	for index, _ := range teCloudAgents.Agents {	
		//if teCloudAgents.Agents[index].Enabled == 1 {
			CloudAgentList[index] = teCloudAgents.Agents[index].AgentID
		//}
	}
	return CloudAgentList
}
//------------------------------------

func main()  {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml") // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")               // optionally look for config in the working directory
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}

	if viper.GetBool("global.debug") {
		Logger("Debugging enabled")
	}

	var teOauthToken = viper.GetString("thousandeyes.oauthToken")
	var teUser = viper.GetString("thousandeyes.user")

	Logger("ThousandEyes Oauth Token: "+teOauthToken)
	Logger("ThousandEyes User: "+teUser)

	Logger("Getting ThousandEyes Account Group Details")
	resp := GetRequest ("https://api.thousandeyes.com/v6/account-groups.json", teOauthToken) 

	teAID := GetAID(resp)
	Logger("Getting ThousandEyes Account Group ID: "+teAID)

	resp = GetRequest ("https://api.thousandeyes.com/v6/agents.json?aid="+teAID+"&agentTypes=ENTERPRISE", teOauthToken)
	EnterpriseAgentList := GetEnterpriseAgents(resp)
	Logger("Enabled Enterprise Agents found: "+strconv.Itoa(len(EnterpriseAgentList)))

	resp = GetRequest ("https://api.thousandeyes.com/v6/agents.json?aid="+teAID+"&agentTypes=ENTERPRISE_CLUSTER", teOauthToken)
	EnterpriseAgentClusterList := GetEnterpriseAgentsCluster(resp)
	Logger("Enabled Enterprise Agent Clusters found: "+strconv.Itoa(len(EnterpriseAgentClusterList)))

	resp = GetRequest ("https://api.thousandeyes.com/v6/agents.json?aid="+teAID+"&agentTypes=CLOUD", teOauthToken)
	CloudAgentList := GetCloudAgents(resp)
	Logger("Used Cloud Agents found: "+strconv.Itoa(len(CloudAgentList)))
}