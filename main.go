package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var version Version
var GitHash string
var BuildTime string
var GoVer string

//change this or set it in Environment variable `BASE_GIT_URL` URL must have 3 instances of %s for code to work
var baseGitUrl string = "https://host.domain.com/api/v4/projects/%s/ref/%s/trigger/pipeline?token=%s&variables[IMAGE_SHA]=%s"

type Version struct {
	GitCommit  string
	ApiVersion string
	GoVersion  string
	BuildDate  string
}

func main() {
	version = Version{GitCommit: GitHash, ApiVersion: "1.0.0", BuildDate: BuildTime, GoVersion: GoVer}
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
		//TODO Filter and figure out which pipeline to trigger
		token := "your_token"     //gitlab pipeline trigger token (may want to make this a lookup or environment)
		project := "your_project" //gitlab project (may want to make this a lookup or environment)
		ref := "main"             //gitlab ref (may want to make this a lookup or environment)
		//TODO end
		triggerPipeline(project, ref, token, hook.EventData.Resources[0].Digest)
		c.JSON(http.StatusCreated, gin.H{
			"created": "true",
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to parse body",
		})
	}
}

func triggerPipeline(project string, ref string, token string, imageSha string) {
	url := fmt.Sprintf(baseGitUrl, project, ref, token)
	log.Printf("url = %s", url)
	resp, err := http.Post(url, "text/plain", nil)
	if err != nil {
		log.Printf("Error posting to API : %s -> %s", imageSha, err)
		return
	}
	if resp.StatusCode != 201 {
		log.Printf("Error posting to API : %s", imageSha)
		log.Printf("Status Code : %d", resp.StatusCode)
	} else {
		fmt.Printf("%s sent to api\n", imageSha)
	}
}
