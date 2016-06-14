package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/plouc/go-gitlab-client"
	"io/ioutil"
	"os"
	"time"
)

type Config struct {
	Host                string `json:"host"`
	ApiPath             string `json:"api_path"`
	Token               string `json:"token"`
	PublishingBaseUrl   string `json:"publishing_base_url"`
	PublishingLogin     string `json:"publishing_login"`
	PublishingPassword  string `json:"publishing_password"`
	FlowdockSourceToken string `json:"flowdock_source_token"`
}

func main() {
	help := flag.Bool("help", false, "Show usage")

	file, e := ioutil.ReadFile("config.json")
	if e != nil {
		fmt.Printf("Config file error: %v\n", e)
		os.Exit(1)
	}

	var config Config
	json.Unmarshal(file, &config)
	fmt.Printf("Results: %+v\n", config)

	gitlab := gogitlab.NewGitlab(config.Host, config.ApiPath, config.Token)

	var method string
	flag.StringVar(&method, "m", "", "Specify method, available methods:\n"+
		"  > -m setupci -id PROJECT_ID\n"+
		"  > -m variables -id PROJECT_ID\n"+
		"  > -m runner -id PROJECT_ID\n"+
	"")

	var id string
	flag.StringVar(&id, "id", "", "Specify project id")

	flag.Usage = func() {
		fmt.Printf("Usage:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *help == true || method == "" {
		flag.Usage()
		return
	}

	startedAt := time.Now()
	defer func() {
		fmt.Printf("processed in %v\n", time.Now().Sub(startedAt))
	}()

	switch method {

	case "setupci":

		if id == "" {
			flag.Usage()
			return
		}

		fmt.Println("Setup project CI...")

		addVariable(gitlab, id, "ORG_GRADLE_PROJECT_publishingBaseUrl", config.PublishingBaseUrl)
		addVariable(gitlab, id, "ORG_GRADLE_PROJECT_publishingLogin", config.PublishingLogin)
		addVariable(gitlab, id, "ORG_GRADLE_PROJECT_publishingPassword", config.PublishingPassword)
		addVariable(gitlab, id, "FLOWDOCK_SOURCE_TOKEN", config.FlowdockSourceToken)
		enableSharedRunners(gitlab, id)

	case "variables":

		if id == "" {
			flag.Usage()
			return
		}

		fmt.Println("Setting project variables...")

		addVariable(gitlab, id, "ORG_GRADLE_PROJECT_publishingBaseUrl", config.PublishingBaseUrl)
		addVariable(gitlab, id, "ORG_GRADLE_PROJECT_publishingLogin", config.PublishingLogin)
		addVariable(gitlab, id, "ORG_GRADLE_PROJECT_publishingPassword", config.PublishingPassword)
		addVariable(gitlab, id, "FLOWDOCK_SOURCE_TOKEN", config.FlowdockSourceToken)

	case "runner":
		if id == "" {
			flag.Usage()
			return
		}

		enableSharedRunners(gitlab, id)
	}
}

func enableSharedRunners(gitlab *gogitlab.Gitlab, id string) (*gogitlab.Project, error) {
	fmt.Printf("Enabling shared runners on project %s", id)

	project := gogitlab.Project{
		SharedRunners: true,
	}

	result, err := gitlab.UpdateProject(id, &project)
	if err != nil {
		fmt.Println(err.Error())
	}

	return result, err
}

func addVariable(gitlab *gogitlab.Gitlab, id string, key string, value string) (*gogitlab.Variable, error) {
	fmt.Printf("Setting project %s variable %s -> %s\n", id, key, value)

	variable := gogitlab.Variable{
		Key:   key,
		Value: value,
	}

	result, err := gitlab.AddProjectVariable(id, &variable)
	if err != nil {
		fmt.Println(err.Error())
	}

	return result, err
}
