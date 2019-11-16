package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

import (
	"github.com/gin-contrib/cors"
	//"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
)

func main() {
	authURL := "https://www.bungie.net/platform/app/oauth/token/"

	//Port the application will listen on
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	//API Key for the Bungie API
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY must be set")
	}

	//Client ID for the Bungie API OAUTH endpoint
	clientID := os.Getenv("CLIENT_ID")
	if clientID == "" {
		log.Fatal("CLIENT_ID must be set")
	}

	//Client Secret for the Bungie API OAUTH endpoint
	clientSecret := os.Getenv("CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatal("CLIENT_SECRET must be set")
	}

	//create router
	r := gin.Default()
	//Setup CORS
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:        false,
		AllowOrigins:           []string{"https://dev.jville.family"},
		AllowMethods:           []string{"OPTIONS", "GET", "POST", "HEAD"},
		AllowHeaders:           []string{"Origin", "Accept", "Content-Type"},
		AllowCredentials:       false,
		ExposeHeaders:          []string{"Content-Length", "Content-Language", "Content-Type"},
		MaxAge:                 60,
		AllowBrowserExtensions: false,
	}))
	
	//Handle base route
	r.GET("/", func(c *gin.Context){
		c.String(http.StatusOK, "Authentication Proxy for the bungie.net api")
	})
	//Proxy authentication request to the Bungie api
	//Adds client secret not available in JS frontend
	r.POST("/authenticate", func(c *gin.Context){
		var authBody AuthBody

		//parse the received body into struct
		err := c.ShouldBind(&authBody)
		if err != nil {
			log.Println("Invalid Post Body sent")
		}

		//Build the request body
		data := url.Values{}
		data.Set("grant_type", authBody.GrantType)
		data.Set("code", authBody.Code)
		data.Set("client_id", clientID)
		data.Set("client_secret", clientSecret)

		//build the request and set required headers
		r, err := http.NewRequest("POST", authURL, strings.NewReader(data.Encode()))
		if err != nil {
			log.Fatal(err)
		}
		r.Header.Set("Content-type", "application/x-www-form-urlencoded")
		r.Header.Set("X-API-KEY", "")

		//execute the request
		client:= &http.Client{}
		resp, err := client.Do(r)
		if err != nil {
			log.Fatal("Error fetching data from bungie.net API: ", err)
		}
		log.Println(resp.StatusCode)

		//handle the response
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Invalid response body received from bungie.net API: ", err)
		}
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	})
	log.Fatal(r.Run(":" + port))
}

type AuthBody struct {
	GrantType string `json:"grantType" form:"grant_type" binding:"required"`
	Code string `json:"code" form:"code" binding:"required"`
}

type BungieAuthResponse struct {
	Error string `json:"error"`
	ErrorDescription string `json:"error_description"`
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}