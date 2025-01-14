package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/javitab/go-web/cli"
	cli_auth "github.com/javitab/go-web/cli/auth"
	dbase "github.com/javitab/go-web/database"
	"github.com/javitab/go-web/router"
	"github.com/joho/godotenv"
)

func StartWebServer() *gin.Engine {
	// Log Server Start Attempt
	dbase.CreateServerStartEvent()

	router := router.AppRouter()

	// Configure the HTTP Server
	httpPort := os.Getenv("HTTP_PORT")

	srv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
		// set timeout due CWE-400 - Potential Slowloris Attack
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start the HTTP server
	log.Println("Starting server on :8080...")
	srv_err := srv.ListenAndServe()

	// Log the server start failure event
	if srv_err != nil {
		dbase.CreateServerStartFailureEvent(srv_err)
		log.Fatal(srv_err)
	}

	return router
}

func WebServerDefaultMode() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	serve := StartWebServer()
	return serve
}

// @title           Go Web API Documentation
// @version         1.0
// @description     This is a sample Go-based CRUD Application Template

// @contact.name   John Avitable
// @contact.url    https://github.com/javitab

// @license.name  GPL-3.0
// @license.url   https://github.com/javitab/go-web/blob/main/LICENSE

// @securityDefinitions.apikey	ApiKeyAuth
// @in	header
// @name Authorization
// @description JWT can be obtained from `login` or `generate_jwt` endpoints. Be sure to include `Bearer` before the JWT

// @host      localhost:8080
// @BasePath  /

func main() {

	fmt.Println(os.Args)

	if len(os.Args) == 1 {
		fmt.Println("No arguments provided")
		cli.PrintHelpText()
		os.Exit(1)
	}

	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error while loading .env file")
	}

	// Initialize the database
	dbase.InitializeDB()

	// ### ###
	// ### ###
	// ### ### Evalaute CLI Arguments
	// ### ###
	// ### ###

	// ### ###
	// ### ### Start Web Server
	// ### ###

	if os.Args[1] == "web" {

		// Check if additional parameters provided
		if len(os.Args) > 2 {

			// Evaluate web server run mode, if no mode, run in release mode
			switch os.Args[2] {
			case "debug":
				_ = StartWebServer()
			default:
				_ = WebServerDefaultMode()
			}
		} else {
			// If no additional parameters, same as default above
			_ = WebServerDefaultMode()
		}

	} else {

		if os.Args[1] == "create_superuser" {
			cli_auth.CLICreateUser()
			return
		}

		// ### ###
		// ### ### Authenticate User and Proceed evaluating CLI Inputs
		// ### ###

		cli_auth.CLIUserLogin()
		// Print info for logged in user
		// helpers.ClearScreen()
		fmt.Println("Utility running as user: " + cli_auth.LoggedInUser.DB.Username)
		if cli_auth.LoggedInUser.IsLDAPUser {
			fmt.Println("User is authenticated with LDAP")
		}

		switch os.Args[1] {
		case "util":
			modes := cli.UtilityMenus[os.Args[2]]
			cli.ExecUtilMenu(modes)
		case "help":

		}

	}

}
