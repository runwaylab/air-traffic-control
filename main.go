package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
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

func main() {
	// Load in the `.env` file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("failed to load env", err)
	}

	// Open a connection to the database
	db, err = sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal("failed to open db connection", err)
	}

	// Build router & define routes
	router := gin.Default()
	router.GET("/commands", GetCommands)
	router.GET("/commands/:commandId", GetSingleCommand)
	router.POST("/commands", CreateCommand)
	router.PUT("/commands/:commandId", UpdateCommand)
	router.DELETE("/commands/:commandId", DeleteCommand)

	// Run the router
	router.Run()
}

func GetCommands(c *gin.Context) {
	query := "SELECT * FROM commands"
	res, err := db.Query(query)
	defer res.Close()
	if err != nil {
		log.Fatal("(GetCommands) db.Query", err)
	}

	commands := []Command{}
	for res.Next() {
		var command Command
		err := res.Scan(&command.Id, &command.Organization, &command.Repository, &command.Name, &command.Data, &command.Created_at, &command.Updated_at)
		if err != nil {
			log.Fatal("(GetCommands) res.Scan", err)
		}
		commands = append(commands, command)
	}

	c.JSON(http.StatusOK, commands)
}

func GetSingleCommand(c *gin.Context) {
	productId := c.Param("productId")
	productId = strings.ReplaceAll(productId, "/", "")
	productIdInt, err := strconv.Atoi(productId)
	if err != nil {
		log.Fatal("(GetSingleCommand) strconv.Atoi", err)
	}

	var command Command
	query := `SELECT * FROM commands WHERE id = ?`
	err = db.QueryRow(query, productIdInt).Scan(&command.Id, &command.Organization, &command.Repository, &command.Name)
	if err != nil {
		log.Fatal("(GetSingleCommand) db.Exec", err)
	}

	c.JSON(http.StatusOK, command)
}

func CreateCommand(c *gin.Context) {
	var newCommand Command
	err := c.BindJSON(&newCommand)
	if err != nil {
		log.Fatal("(CreateCommand) c.BindJSON", err)
	}

	query := `INSERT INTO commands (id, organization, repository, name) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, newCommand.Id, newCommand.Organization, newCommand.Repository, newCommand.Name)
	if err != nil {
		log.Fatal("(CreateCommand) db.Exec", err)
	}

	_, err = res.LastInsertId()

	if err != nil {
		log.Fatal("(CreateCommand) res.LastInsertId", err)
	}

	c.JSON(http.StatusOK, newCommand)
}

func UpdateCommand(c *gin.Context) {
	var updates Command
	err := c.BindJSON(&updates)
	if err != nil {
		log.Fatal("(UpdateCommand) c.BindJSON", err)
	}

	commandId := c.Param("commandId")
	commandId = strings.ReplaceAll(commandId, "/", "")

	query := `UPDATE commands SET name = ?, repository = ? WHERE id = ?`
	_, err = db.Exec(query, updates.Name, updates.Repository, commandId)
	if err != nil {
		log.Fatal("(UpdateCommand) db.Exec", err)
	}

	c.Status(http.StatusOK)
}

func DeleteCommand(c *gin.Context) {
	commandId := c.Param("commandId")

	commandId = strings.ReplaceAll(commandId, "/", "")

	query := `DELETE FROM commands WHERE id = ?`
	_, err := db.Exec(query, commandId)
	if err != nil {
		log.Fatal("(DeleteCommand) db.Exec", err)
	}

	c.Status(http.StatusOK)
}