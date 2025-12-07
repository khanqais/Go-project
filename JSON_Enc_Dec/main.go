package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Todo struct {
	UserId    int    `json:"userId"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func main() {
	url := "https://jsonplaceholder.typicode.com/todos/1/"

	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		// bodyBytes, err := io.ReadAll(res.Body)
		// if err != nil {
		// 	panic(err)
		// }
		todoItem := Todo{}
		// json.Unmarshal(bodyBytes, &todoItem)
		// fmt.Printf("Data from API: %+v", todoItem)
		decoder := json.NewDecoder(res.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&todoItem); err != nil 
		{
			log.Fatal("Decoder Error", err)
		}
		fmt.Printf("Data from API: %+v", todoItem)
	}
}
