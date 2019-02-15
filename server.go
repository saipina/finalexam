package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Customer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

var Customers []Customer
var DB *sql.DB

func main() {
	//!Database Connection
	//set POSTGRES_DB_URL=postgres://etyvqaeb:JpaMcN_E6eM0XoOIPPcm31A5eu76E57K@baasu.db.elephantsql.com:5432/etyvqaeb
	var err error
	DB, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	//!Gin Here
	r := gin.Default()
	//api := r.Group("/api")
	r.Use(loginMiddleware)
	r.POST("/customers", postCustHandler)
	r.GET("/customers/:id", getCustByIDHandler)
	r.GET("/customers", getCustAllHandler)
	r.PUT("/customers/:id", putCustByIDHandler)
	r.DELETE("/customers/:id", delCustByIDHandler)
	r.Run(":2019")

}

func delCustByIDHandler(c *gin.Context) {
	id := c.Param("id")

	stmt, err := DB.Prepare("DELETE FROM customers WHERE id=$1")
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	_, err = stmt.Exec(id)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(200, gin.H{
		"message": "customer deleted",
	})

}

func putCustByIDHandler(c *gin.Context) {
	id := c.Param("id")
	var newCustomer Customer

	err := c.ShouldBindJSON(&newCustomer)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	stmt, err := DB.Prepare("UPDATE customers SET name=$4,email=$3, status=$2 where id=$1")
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	_, err = stmt.Exec(id, newCustomer.Status, newCustomer.Email, newCustomer.Name)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	//!Query By ID
	stmt2, err := DB.Prepare("SELECT id, name, email, status FROM customers where id = $1")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	row := stmt2.QueryRow(id)
	var updCus Customer
	err = row.Scan(&updCus.ID, &updCus.Name, &updCus.Email, &updCus.Status)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	//!
	c.JSON(200, updCus)

}

func getCustAllHandler(c *gin.Context) {
	stmt, err := DB.Prepare("SELECT id, name, email, status FROM customers")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	rows, err := stmt.Query()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	var customers []Customer
	for rows.Next() {
		var custumer Customer
		err := rows.Scan(&custumer.ID, &custumer.Name, &custumer.Email, &custumer.Status)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		customers = append(customers, custumer)

	}

	c.JSON(200, customers)

}

func getCustByIDHandler(c *gin.Context) {
	id := c.Param("id")
	stmt, err := DB.Prepare("SELECT id, name, email,  status FROM customers where id = $1")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	row := stmt.QueryRow(id)
	var customer Customer
	err = row.Scan(&customer.ID, &customer.Name, &customer.Email, &customer.Status)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	//!
	c.JSON(200, customer)

}

func postCustHandler(c *gin.Context) {
	var newCustomer Customer
	fmt.Println("post it")
	err := c.ShouldBindJSON(&newCustomer)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	//!
	//fmt.Println(newItem)

	row := DB.QueryRow("INSERT INTO customers (name, email, status) values ($1, $2, $3) RETURNING id", newCustomer.Name, newCustomer.Email, newCustomer.Status)
	var id int
	err = row.Scan(&id)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	newCustomer.ID = id
	c.JSON(201, newCustomer)

}
func loginMiddleware(c *gin.Context) {
	authKey := c.GetHeader("Authorization")
	if authKey != "token2019" {
		c.JSON(http.StatusUnauthorized, "Unauthorization")
		c.Abort()
		return
	}
	c.Next()
	log.Println("ending middleware")
}
