package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Games struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Game     string             `bson:"game,omitempty"`
	CharList []string           `bson:"charList,omitempty"`
}

type DegreesObject struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Name    string             `bson:"name,omitempty"`
	Degrees int                `bson:"degrees,omitempty"`
	Links   []LinkObject       `bson:"links,omitempty"`
}

type LinkObject struct {
	Character string `json:"character"`
	Game      string `json:"game"`
	Year      string `json:"year"`
}

type CharacterInput struct {
	Name string `json:"name"`
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func createNewArticle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reqBody, _ := ioutil.ReadAll(r.Body)
	var input CharacterInput

	println("reqBody1 = " + string(reqBody))
	json.Unmarshal(reqBody, &input)
	println("input = " + string(input.Name))

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://ryu:dBwlVrY7OVU93Y91@cluster0.he74l.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(databases)

	quickstartDatabase := client.Database("six-degrees")
	degreesCollection := quickstartDatabase.Collection("degrees")

	regex := bson.M{"$regex": primitive.Regex{Pattern: "^" + input.Name + "$", Options: "i"}}

	filterCursor, err := degreesCollection.Find(ctx, bson.M{"name": regex})
	if err != nil {
		log.Fatal(err)
	}
	client.Disconnect(ctx)

	var results []DegreesObject
	if err = filterCursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}
	fmt.Println(results)

	for _, element := range results {
		fmt.Println(element.Degrees)
		json.NewEncoder(w).Encode(element)
	}

}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/article", createNewArticle).Methods("POST")

	myRouter.Headers("Access-Control-Allow-Origin", "*")

	port := os.Getenv("PORT")
	if len(port) <= 0 {
		port = "10000"
	}
	println("porta = " + port)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(myRouter)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func main() {
	handleRequests()
}
