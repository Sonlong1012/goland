package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

var client *firestore.Client

func initializeFirebase() {
	// Replace the path with the actual path to your Firebase Admin SDK JSON file
	opt := option.WithCredentialsFile("config/a.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatal(err)
	}

	client, err = app.Firestore(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
func defaultPage(c *gin.Context) {

	c.JSON(200, gin.H{
		"Hello": "This is init page",
	})
}
func getAllDocumentsName(c *gin.Context) {
	collection := client.Collection("Items")

	// Get all documents from the collection
	docs, err := collection.Documents(context.Background()).GetAll()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve documents"})
		return
	}

	var names []string

	for _, doc := range docs {
		// Extract the value of the "Name" field from each document
		nameValue, found := doc.Data()["Name"]
		if !found {
			// Field not found in the document, handle accordingly
			continue
		}

		// Assert and convert the field value to string
		name, ok := nameValue.(string)
		if !ok {
			// Field value is not a string, handle accordingly
			continue
		}

		// Append the name to the slice
		names = append(names, name)
	}

	c.JSON(200, names)
}
func getAllDocuments(c *gin.Context) {
	collection := client.Collection("Items")

	// Get all documents from the collection
	docs, err := collection.Documents(context.Background()).GetAll()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve documents"})
		return
	}

	var result []map[string]interface{}

	for _, doc := range docs {
		data := doc.Data()
		result = append(result, data)
	}

	c.JSON(200, result) //trả dữ liệu về
}
func getDocumentsBaseOnID(c *gin.Context) {
	idStr := c.Param("id")
	log.Println("Updating Name for document with ID:", idStr)

	// Convert the string to an integer
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("Error converting ID to integer: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Get the document reference
	iter := client.Collection("Items").Where("ID", "==", id).Documents(context.Background())
	doc, err := iter.Next()
	if err != nil {
		log.Printf("Error retrieving document: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Extract data from the document
	data := doc.Data()

	c.JSON(http.StatusOK, data)
}
func addDocument(c *gin.Context) {
	// Replace "your_collection" with the name of your Firestore collection
	var newItem Item
	if err := c.BindJSON(&newItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("bind JSON success")
	_, _, err := client.Collection("Items").Add(context.Background(), newItem)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
		return
	}
	c.JSON(200, gin.H{"success": "success"})
	fmt.Println("add data success")
	c.Status(http.StatusCreated)
}
func updateDocumentsBaseOnID(c *gin.Context) {

	idStr := c.Param("id")
	log.Println("Updating Name for document with ID:", idStr)

	// Convert the string to an integer
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("Error converting ID to integer: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Get the document reference
	iter := client.Collection("Items").Where("ID", "==", id).Documents(context.Background())
	doc, err := iter.Next()
	if err != nil {
		log.Printf("Error retrieving document: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Bind the updated data from the request
	var updateData struct {
		Name string `json:"name" firestore:"Name"`
	}

	if err := c.BindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the 'Name' field in the document
	_, err = doc.Ref.Update(context.Background(), []firestore.Update{
		{Path: "Name", Value: updateData.Name},
	})

	if err != nil {
		log.Printf("Error updating document: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"update": "success"})
}
func deleteDocumentsBaseOnID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("Error converting ID to integer: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	log.Println("Deleting document with ID:", id)

	iter := client.Collection("Items").Where("ID", "==", id).Documents(context.Background())
	doc, err := iter.Next()
	if err != nil {
		log.Printf("Error retrieving document: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	_, err = doc.Ref.Delete(context.Background())
	if err != nil {
		log.Printf("Failed to delete document: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"delete": "success"})

}

func main() {
	initializeFirebase()

	// Create a new Gin router
	router := gin.Default()

	// Define a route to get all documents
	router.GET("/", defaultPage)

	// router.GET("/getall", getAllDocuments)
	apiGroup := router.Group("/api")
	// Routes under "/api" group
	apiGroup.GET("/getall", getAllDocuments)

	apiGroup.GET("/getname", getAllDocumentsName)

	apiGroup.GET("/get/:id", getDocumentsBaseOnID)

	router.PUT("/update/:id", updateDocumentsBaseOnID)

	router.DELETE("/delete/:id", deleteDocumentsBaseOnID)

	router.POST("/add", addDocFument)
	// Run the server on port 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run("0.0.0.0:" + port)
}

type Item struct {
	ID          int64  `json:"id" firestore:"ID"`
	Name        string `json:"name" firestore:"Name"`
	Description string `json:"description" firestore:"Description"`
}
