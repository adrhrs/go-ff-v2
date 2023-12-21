package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"gopkg.in/yaml.v3"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func initFlag() {
	ffclient.Init(ffclient.Config{
		PollingInterval: 10 * time.Second,
		Retriever: &fileretriever.Retriever{
			Path: "flag.yaml",
		},
	})

}

type ApiResponse struct {
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	cohort := r.URL.Query().Get("cohort")

	user := ffcontext.
		NewEvaluationContextBuilder(userID).
		AddCustom("cohort", cohort).
		Build()
	allFlagsState := ffclient.AllFlagsState(user)

	apiResponse := ApiResponse{
		Msg:  "success",
		Data: allFlagsState,
	}

	jsonData, err := json.Marshal(apiResponse)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func writeToFile(fileName, content string) error {
	err := os.WriteFile(fileName, []byte(content), 0644)
	if err != nil {
		return err
	}
	return nil
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	// Read the JSON data from the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Parse the JSON data into a map
	var jsonData map[string]interface{}
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		http.Error(w, "Error parsing JSON data", http.StatusBadRequest)
		return
	}

	// Convert JSON to YAML
	yamlData, err := yaml.Marshal(jsonData)
	if err != nil {
		http.Error(w, "Error converting JSON to YAML", http.StatusInternalServerError)
		return
	}

	// Write YAML content to the file, overwriting existing content
	fileName := "flag.yaml"
	err = writeToFile(fileName, string(yamlData)+"\n")
	if err != nil {
		http.Error(w, "Error writing to file", http.StatusInternalServerError)
		return
	}

	// Create the API response
	apiResponse := ApiResponse{
		Msg:  "success",
		Data: jsonData,
	}

	// Convert the API response to JSON
	jsonResponse, err := json.Marshal(apiResponse)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set the content type and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

func main() {
	initFlag()

	router := mux.NewRouter()

	corsOptions := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)

	router.HandleFunc("/api/get", getHandler).Methods("GET")
	router.HandleFunc("/api/writeyaml", postHandler).Methods("POST")

	port := 8080
	fmt.Printf("nexus Server started on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), corsOptions(router))

}
