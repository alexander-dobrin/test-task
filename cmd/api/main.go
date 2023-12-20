package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GetCourseResponse struct {
	Symbol             string `json:"symbol"`
	PriceChangePercent string `json:"priceChangePercent"`
	LastPrice          string `json:"lastPrice"`
}

func handleGetCourse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response, err := http.Get("https://api.binance.com/api/v3/ticker/24hr?symbol=BTCUSDT")

	if err != nil {
		fmt.Println("Api error")
		return
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var result GetCourseResponse

	err = json.Unmarshal(body, &result)

	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	fmt.Println(result)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
	}

	dbPassword := os.Getenv("DB_PASSWORD")

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	fmt.Print(fmt.Sprintf("mongodb+srv://devasted:%s@cluster0.f7qdrod.mongodb.net/?retryWrites=true&w=majority", dbPassword))
	opts := options.Client().ApplyURI(fmt.Sprintf("mongodb+srv://devasted:%s@cluster0.f7qdrod.mongodb.net/?retryWrites=true&w=majority", dbPassword)).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		panic(err)
	}

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)

	if err != nil {
		port = 3333
	}

	fmt.Println("Starting GO API")

	http.HandleFunc("/api/GetCourse", handleGetCourse)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	if err != nil {
		fmt.Println(err.Error())
	}
}
