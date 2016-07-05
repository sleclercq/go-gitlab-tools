package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/plouc/go-gitlab-client"
	"io/ioutil"
	"os"
	"strconv"
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
		"  > -m create -name PROJECT_NAME -namespace_id PROJECT_NAMESPACE\n"+
		"")

	var id string
	flag.StringVar(&id, "id", "", "Specify project id")

	var name string
	flag.StringVar(&name, "name", "", "Specify project name")

	var namespaceId string
	flag.StringVar(&namespaceId, "namespace_id", "", "Specify project namespace id")

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

	case "create":
		if name == "" || namespaceId == "" {
			flag.Usage()
			return
		}

		namespaceIdInt, _ := strconv.Atoi(namespaceId)
		project, err := createProject(gitlab, name, namespaceIdInt)

		if err != nil {
			return
		}

		setupCi(gitlab, strconv.Itoa(project.Id), config)

	case "setupci":

		if id == "" {
			flag.Usage()
			return
		}

		fmt.Println("Setup project CI...")

		setupCi(gitlab, id, config)

	case "variables":

		if id == "" {
			flag.Usage()
			return
		}

		setupProjectVariables(gitlab, id, config)

	case "runner":
		if id == "" {
			flag.Usage()
			return
		}

		enableSharedRunners(gitlab, id)
	}
}

func createProject(gitlab *gogitlab.Gitlab, name string, namespaceId int) (*gogitlab.Project, error) {
	fmt.Printf("Creating project with name %v and namespace %v\n", name, namespaceId)

	project := gogitlab.Project{
		Name:                 name,
		NamespaceId:          namespaceId,
		MergeRequestsEnabled: true,
		PublicBuilds:         false,
	}

	result, err := gitlab.CreateProject(&project)

	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("Created project with id %v and url %v\n", result.Id, result.WebUrl)

	return result, err
}

func setupCi(gitlab *gogitlab.Gitlab, id string, config Config) (*gogitlab.Project, error) {
	err := setupProjectVariables(gitlab, id, config)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	result, err := enableSharedRunners(gitlab, id)
	if err != nil {
		fmt.Println(err.Error())
		return result, err
	}

	return result, err
}

func setupProjectVariables(gitlab *gogitlab.Gitlab, id string, config Config) error {
	fmt.Println("Setting project variables...")

	_, err := addVariable(gitlab, id, "ORG_GRADLE_PROJECT_publishingBaseUrl", config.PublishingBaseUrl)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	_, err = addVariable(gitlab, id, "ORG_GRADLE_PROJECT_publishingLogin", config.PublishingLogin)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	_, err = addVariable(gitlab, id, "ORG_GRADLE_PROJECT_publishingPassword", config.PublishingPassword)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	_, err = addVariable(gitlab, id, "FLOWDOCK_SOURCE_TOKEN", config.FlowdockSourceToken)
	return err
}

func enableSharedRunners(gitlab *gogitlab.Gitlab, id string) (*gogitlab.Project, error) {
	fmt.Printf("Enabling shared runners on project %v\n", id)

	project := gogitlab.Project{
		BuildsEnabled: true,
		PublicBuilds:  false,
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
