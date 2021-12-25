package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var version Version
var GoVer string

//change this or set it in Environment variable `BASE_GIT_URL` URL must have 3 instances of %s for code to work
var baseGitUrl = "https://git.home.starkenberg.net/api/v4/projects/%s/ref/%s/trigger/pipeline?token=%s&variables[IMAGE_TAG]=%s"

type Version struct {
	ApiVersion string
	GoVersion  string
}

type ProjectMetaData struct {
	Token     string `json:"token"`
	ProjectID string `json:"projectId"`
	Ref       string `json:"ref"`
}

func main() {
	version = Version{ApiVersion: "1.0.1", GoVersion: GoVer}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println(sig)
		fmt.Println("Exiting")
		os.Exit(0)
	}()
	envUrl, urlExists := os.LookupEnv("BASE_GIT_URL")
	if urlExists {
		baseGitUrl = envUrl
	}
	handleRequests()
}

func handleRequests() {
	r := gin.Default()
	r.GET("/", homePage)
	r.POST("/image", createEvent)
	log.Fatal(r.Run(":8080"))
}

func homePage(c *gin.Context) {
	fmt.Println("Endpoint Hit: homePage")
	c.JSON(http.StatusOK, &version)
}

func createEvent(c *gin.Context) {
	var hook HarborHook
	if c.ShouldBind(&hook) == nil {
		log.Printf("Hook : %+v", hook)
		project, err := readMetaData(hook.EventData.Repository.Name)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		}
		triggerPipeline(project, hook.EventData.Resources[0].Tag)
		c.JSON(http.StatusCreated, gin.H{
			"created": "true",
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to parse body",
		})
	}
}

func triggerPipeline(project ProjectMetaData, imageTag string) {
	url := fmt.Sprintf(baseGitUrl, project.ProjectID, project.Ref, project.Token, imageTag)
	log.Printf("url = %s", url)
	resp, err := http.Post(url, "text/plain", nil)
	if err != nil {
		log.Printf("Error posting to API : %s -> %s", imageTag, err)
		return
	}
	if resp.StatusCode != 201 {
		log.Printf("Error posting to API : %s", imageTag)
		log.Printf("Status Code : %d", resp.StatusCode)
	} else {
		fmt.Printf("%s sent to api\n", imageTag)
	}
}

func readMetaData(serviceName string) (data ProjectMetaData, err error) {
	file, err := ioutil.ReadFile(fmt.Sprintf("/etc/secrets/%s", serviceName))
	if err != nil {
		log.Printf("Error reading %s secrets file : %s", serviceName, err)
		return
	}
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		log.Printf("Error parsing %s secrets file : %s", serviceName, err)
	}
	return
}
