package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/runwayapp/air-traffic-control/internal/middlewares"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var db *sql.DB

type Command struct {
	Id           string
	Organization string
	Repository   string
	Name         string
	Data         string
	Created_at   string
	Updated_at   string
}

type CommandResponse struct {
	Id           string                 `json:"id"`
	Organization string                 `json:"organization"`
	Repository   string                 `json:"repository"`
	Name         string                 `json:"name"`
	Data         map[string]interface{} `json:"data"`
	Created_at   string                 `json:"created_at"`
	Updated_at   string                 `json:"updated_at"`
}

func main() {
	var err error
	// Load in the `.env` file in development
	if os.Getenv("ENV") != "production" {
		err = godotenv.Load()
		if err != nil {
			log.Fatal("failed to load env", err)
		}
	}

	// Open a connection to the database
	db, err = sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal("failed to open db connection", err)
	}

	// Build router & define routes
	router := gin.Default()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	router.Use(gin.Recovery())

	protected := router.Group("/")
	protected.Use(middlewares.JwtAuthMiddleware())

	protected.GET("/:org/:repo/commands", GetRepoCommands)
	protected.GET("/:org/:repo/commands/:commandId", GetSingleCommand)
	protected.POST("/:org/:repo/commands", CreateCommand)
	protected.PUT("/:org/:repo/commands/:commandId", UpdateCommand)
	protected.DELETE("/:org/:repo/commands/:commandId", DeleteCommand)

	// Run the router
	router.Run()
}

func GetRepoCommands(c *gin.Context) {
	org := c.Param("org")
	org = strings.ReplaceAll(org, "/", "")
	repo := c.Param("repo")
	repo = strings.ReplaceAll(repo, "/", "")

	query := `SELECT * FROM commands WHERE organization = ? AND repository = ?`
	res, err := db.Query(query, org, repo)
	defer res.Close()
	if err != nil {
		msg, _ := fmt.Printf("(GetCommands) db.Query %s", err)
		panic(msg)
	}

	commands := []CommandResponse{}
	for res.Next() {
		var commandResponse CommandResponse
		var command Command
		err := res.Scan(&command.Id, &command.Organization, &command.Repository, &command.Name, &command.Data, &command.Created_at, &command.Updated_at)
		if err != nil {
			msg, _ := fmt.Printf("(GetCommands) res.Scan %s", err)
			panic(msg)
		}

		// ensure the data is valid json before appending
		var data map[string]interface{}
		err = json.Unmarshal([]byte(command.Data), &data)
		if err != nil {
			msg, _ := fmt.Printf("(GetCommands) json.Unmarshal %s", err)
			panic(msg)
		}

		commandResponse.Id = command.Id
		commandResponse.Organization = command.Organization
		commandResponse.Repository = command.Repository
		commandResponse.Name = command.Name
		commandResponse.Created_at = command.Created_at
		commandResponse.Updated_at = command.Updated_at
		commandResponse.Data = data

		commands = append(commands, commandResponse)
	}

	c.JSON(http.StatusOK, commands)
}

func GetSingleCommand(c *gin.Context) {
	org := c.Param("org")
	org = strings.ReplaceAll(org, "/", "")
	repo := c.Param("repo")
	repo = strings.ReplaceAll(repo, "/", "")
	commandId := c.Param("commandId")
	commandId = strings.ReplaceAll(commandId, "/", "")

	var commandResponse CommandResponse
	var command Command
	query := `SELECT * FROM commands WHERE id = ? AND organization = ? AND repository = ?`
	err := db.QueryRow(query, commandId, org, repo).Scan(&command.Id, &command.Organization, &command.Repository, &command.Name, &command.Data, &command.Created_at, &command.Updated_at)
	if err != nil {
		msg, _ := fmt.Printf("(GetSingleCommand) db.Exec %s", err)
		panic(msg)
	}

	// ensure the data is valid json before appending
	var data map[string]interface{}
	err = json.Unmarshal([]byte(command.Data), &data)
	if err != nil {
		msg, _ := fmt.Printf("(GetCommands) json.Unmarshal %s", err)
		panic(msg)
	}

	commandResponse.Id = command.Id
	commandResponse.Organization = command.Organization
	commandResponse.Repository = command.Repository
	commandResponse.Name = command.Name
	commandResponse.Created_at = command.Created_at
	commandResponse.Updated_at = command.Updated_at
	commandResponse.Data = data

	c.JSON(http.StatusOK, commandResponse)
}

func CreateCommand(c *gin.Context) {
	id := uuid.New().String()

	org := c.Param("org")
	org = strings.ReplaceAll(org, "/", "")
	repo := c.Param("repo")
	repo = strings.ReplaceAll(repo, "/", "")

	var newCommand Command
	err := c.BindJSON(&newCommand)
	if err != nil {
		msg, _ := fmt.Printf("(CreateCommand) c.BindJSON %s", err)
		panic(msg)
	}

	// add values to the new command
	newCommand.Id = id
	newCommand.Organization = org
	newCommand.Repository = repo

	// check required params
	if newCommand.Organization == "" || newCommand.Repository == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization and repository are required params"})
		return
	}

	// Check all required inputs
	if newCommand.Name == "" || newCommand.Data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization, repository, name, and data are required"})
		return
	}

	query := `INSERT INTO commands (id, organization, repository, name, data) VALUES (?, ?, ?, ?, ?)`
	res, err := db.Exec(query, newCommand.Id, newCommand.Organization, newCommand.Repository, newCommand.Name, newCommand.Data)
	if err != nil {
		msg, _ := fmt.Printf("(CreateCommand) db.Exec %s", err)
		panic(msg)
	}

	_, err = res.LastInsertId()

	if err != nil {
		msg, _ := fmt.Printf("(CreateCommand) res.LastInsertId %s", err)
		panic(msg)
	}

	var commandResponse CommandResponse

	// ensure the data is valid json before appending
	var data map[string]interface{}
	err = json.Unmarshal([]byte(newCommand.Data), &data)
	if err != nil {
		msg, _ := fmt.Printf("(GetCommands) json.Unmarshal %s", err)
		panic(msg)
	}

	commandResponse.Id = newCommand.Id
	commandResponse.Organization = newCommand.Organization
	commandResponse.Repository = newCommand.Repository
	commandResponse.Name = newCommand.Name
	commandResponse.Created_at = newCommand.Created_at
	commandResponse.Updated_at = newCommand.Updated_at
	commandResponse.Data = data

	c.JSON(http.StatusOK, commandResponse)
}

func UpdateCommand(c *gin.Context) {
	var updates Command
	err := c.BindJSON(&updates)
	if err != nil {
		msg, _ := fmt.Printf("(UpdateCommand) c.BindJSON %s", err)
		panic(msg)
	}

	org := c.Param("org")
	org = strings.ReplaceAll(org, "/", "")
	repo := c.Param("repo")
	repo = strings.ReplaceAll(repo, "/", "")
	commandId := c.Param("commandId")
	commandId = strings.ReplaceAll(commandId, "/", "")

	// check all required inputs
	if updates.Name == "" || updates.Data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and data are required"})
		return
	}

	query := `UPDATE commands SET name = ?, data = ? WHERE id = ? AND organization = ? AND repository = ?`
	result, err := db.Exec(query, updates.Name, updates.Data, commandId, org, repo)
	if err != nil {
		msg, _ := fmt.Printf("(UpdateCommand) db.Exec %s", err)
		panic(msg)
	}

	// if no rows were affected, return an error
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		msg, _ := fmt.Printf("(DeleteCommand) result.RowsAffected %s", err)
		panic(msg)
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "command not found"})
		return
	}

	c.Status(http.StatusOK)
}

func DeleteCommand(c *gin.Context) {
	org := c.Param("org")
	org = strings.ReplaceAll(org, "/", "")
	repo := c.Param("repo")
	repo = strings.ReplaceAll(repo, "/", "")
	commandId := c.Param("commandId")
	commandId = strings.ReplaceAll(commandId, "/", "")

	// if an org or repo is not provided, return an error
	if org == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "org and repo are required"})
		return
	}

	// if a commandId is not provided, return an error
	if commandId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "commandId is required"})
		return
	}

	query := `DELETE FROM commands WHERE id = ? AND organization = ? AND repository = ?`
	result, err := db.Exec(query, commandId, org, repo)
	if err != nil {
		msg, _ := fmt.Printf("(DeleteCommand) db.Exec %s", err)
		panic(msg)
	}

	// if no rows were affected, return an error
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		msg, _ := fmt.Printf("(DeleteCommand) result.RowsAffected %s", err)
		panic(msg)
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "command not found"})
		return
	}

	c.Status(http.StatusOK)
}
