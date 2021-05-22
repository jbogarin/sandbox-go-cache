package main

import (
	"crypto/tls"
	"encoding/xml"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/patrickmn/go-cache"
)

var myCache *cache.Cache

var client *resty.Client

const URL = "https://<FQDN>/getxml?location=/Status/RoomAnalytics/PeoplePresence"
const username = "<USERNAME>"
const password = "<PASSWORD>"

type PeoplePresenceResponse struct {
	XMLName       xml.Name      `xml:"Status"`
	RoomAnalytics RoomAnalytics `xml:"RoomAnalytics"`
}

type RoomAnalytics struct {
	XMLName        xml.Name `xml:"RoomAnalytics"`
	PeoplePresence string   `xml:"PeoplePresence"`
}

func getPeoplePresence() string {
	client = resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	client.SetBasicAuth(username, password)

	var peoplePresence *PeoplePresenceResponse
	resp, err := client.R().
		SetResult(&PeoplePresenceResponse{}).
		ExpectContentType("application/xml").
		Get(URL)

	if err != nil {
		log.Fatal(err)
	}

	peoplePresence = resp.Result().(*PeoplePresenceResponse)

	log.Println("PeoplePresence --> ", peoplePresence.RoomAnalytics.PeoplePresence)
	return peoplePresence.RoomAnalytics.PeoplePresence
}

func main() {
	r := gin.Default()

	myCache = cache.New(30*time.Second, time.Minute)

	peoplePresence := getPeoplePresence()
	myCache.Set("peoplePresence", peoplePresence, cache.DefaultExpiration)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/cache", func(c *gin.Context) {
		peoplePresence, found := myCache.Get("peoplePresence")
		if found {
			log.Println("Using cache")
			c.JSON(200, gin.H{
				"message": peoplePresence,
			})
		} else {
			log.Println("Cache expired")
			peoplePresence := getPeoplePresence()
			myCache.Set("peoplePresence", peoplePresence, cache.DefaultExpiration)
			c.JSON(404, gin.H{
				"message": peoplePresence,
			})
		}
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
