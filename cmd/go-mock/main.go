// Package gomock provides a simple in-memory JSON-based mock server for testing and development.
// It allows creating a lightweight HTTP server that can handle CRUD operations on JSON collections.
//
// The server supports the following HTTP methods:
//   - GET: Retrieve all items in a collection or a specific item by ID
//   - POST: Create a new item in a collection
//   - PUT: Update an existing item in a collection
//   - DELETE: Remove an item from a collection
//
// Usage:
//
//	go run main.go -db=mydata.json -port=8080
//
// Flags:
//
//	-db string
//	  	Path to the JSON database file (default "db.json")
//	-port int
//	  	Port number for the server to listen on (default 3000)
package gomock

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Server represents the mock JSON server with data management capabilities.
// It maintains an in-memory collection of data and provides methods to interact with it.
type Server struct {

	// data stores collections as a map where each key is a collection name
	// and the value is a slice of items (represented as map[string]interface{})
	data map[string][]interface{}

	// filename is the path to the JSON file used for persistent storage
	filename string

	// logger is used for logging server activities and requests
	logger *log.Logger
}

// NewServer initializes a new Server instance by loading data from a JSON file.
//
// Parameters:
//   - filename: Path to the JSON database file
//
// Returns:
//   - *Server: Configured server instance
//   - error: Any error encountered during initialization
func NewServer(filename string) (*Server, error) {
	server := &Server{
		data:     make(map[string][]interface{}),
		filename: filename,
		logger:   log.New(os.Stdout, "GO-MOCK: ", log.Ldate|log.Ltime|log.Lshortfile),
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)
	if err := json.Unmarshal(byteValue, &server.data); err != nil {
		return nil, err
	}

	return server, nil
}

// saveData writes the current in-memory data to the JSON file.
//
// Returns:
//   - error: Any error encountered while saving data
func (s *Server) saveData() error {
	file, err := json.MarshalIndent(s.data, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(s.filename, file, 0644)
}

// Run starts the mock server with command-line configurable options.
//
// Flags:
//   - db: JSON database file path (default: "db.json")
//   - port: Server listening port (default: 3000)
//
// Returns:
//   - error: Any error encountered while starting the server
func Run() error {
	filename := flag.String("db", "db.json", "JSON database file")
	port := flag.Int("port", 3000, "Server port")
	flag.Parse()

	server, err := NewServer(*filename)
	if err != nil {
		return fmt.Errorf("error initializing server: %v", err)
	}

	http.HandleFunc("/", server.handleCollection)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Mock Server running on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// handleCollection is the main HTTP handler for processing requests to different collections.
// It routes requests based on HTTP method (GET, POST, PUT, DELETE) to appropriate handler functions.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request
func (s *Server) handleCollection(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	collection := parts[0]

	var status int
	defer func() { s.logRequest(r, collection, status) }()

	switch r.Method {
	case http.MethodGet:
		status = s.handleGet(w, r, collection, parts)
	case http.MethodPost:
		status = s.handlePost(w, r, collection)
	case http.MethodPut:
		status = s.handlePut(w, r, collection, parts)
	case http.MethodDelete:
		status = s.handleDelete(w, r, collection, parts)
	default:
		status = http.StatusMethodNotAllowed
		http.Error(w, "Method not allowed", status)
	}
}

// logRequest logs details about each incoming HTTP request.
//
// Parameters:
//   - r: HTTP request
//   - collection: Target collection name
//   - status: HTTP response status code
func (s *Server) logRequest(r *http.Request, collection string, status int) {
	s.logger.Printf("%s %s /%s - Status: %d", r.Method, r.RemoteAddr, collection, status)
}

// findItemIndex locates the index of an item in a collection by its ID.
//
// Parameters:
//   - collection: Name of the collection to search
//   - id: Numeric identifier of the item
//
// Returns:
//   - int: Index of the item (or -1 if not found)
//   - bool: Whether the item was found
func (s *Server) findItemIndex(collection string, id int) (int, bool) {
	items := s.data[collection]
	for i, item := range items {
		itemMap := item.(map[string]interface{})
		if int(itemMap["id"].(float64)) == id {
			return i, true
		}
	}
	return -1, false
}

// handleGet processes GET requests for retrieving collection items.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request
//   - collection: Target collection name
//   - parts: URL path segments
//
// Returns:
//   - int: HTTP response status code
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request, collection string, parts []string) int {
	items, exists := s.data[collection]
	if !exists {
		http.Error(w, "Collection not found", http.StatusNotFound)
		return http.StatusNotFound
	}

	// Handle specific item retrieval
	if len(parts) > 1 {
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return http.StatusBadRequest
		}

		for _, item := range items {
			itemMap := item.(map[string]interface{})
			if int(itemMap["id"].(float64)) == id {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(itemMap)
				return http.StatusOK
			}
		}
		http.Error(w, "Item not found", http.StatusNotFound)
		return http.StatusNotFound
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
	return http.StatusOK
}

// handlePost processes POST requests for creating new items in a collection.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request
//   - collection: Target collection name
//
// Returns:
//   - int: HTTP response status code
func (s *Server) handlePost(w http.ResponseWriter, r *http.Request, collection string) int {
	var newItem map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return http.StatusBadRequest
	}

	if _, hasID := newItem["id"]; !hasID {
		newItem["id"] = len(s.data[collection]) + 1
	}

	s.data[collection] = append(s.data[collection], newItem)

	if err := s.saveData(); err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return http.StatusInternalServerError
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newItem)
	return http.StatusCreated
}

// handlePut processes PUT requests for updating existing items in a collection.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request
//   - collection: Target collection name
//   - parts: URL path segments
//
// Returns:
//   - int: HTTP response status code
func (s *Server) handlePut(w http.ResponseWriter, r *http.Request, collection string, parts []string) int {
	if len(parts) < 2 {
		http.Error(w, "ID required", http.StatusBadRequest)
		return http.StatusBadRequest
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return http.StatusBadRequest
	}

	var updatedItem map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updatedItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return http.StatusBadRequest
	}

	idx, found := s.findItemIndex(collection, id)
	if !found {
		http.Error(w, "Item not found", http.StatusNotFound)
		return http.StatusNotFound
	}

	updatedItem["id"] = id
	s.data[collection][idx] = updatedItem

	if err := s.saveData(); err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return http.StatusInternalServerError
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedItem)
	return http.StatusOK
}

// handleDelete processes DELETE requests for removing items from a collection.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request
//   - collection: Target collection name
//   - parts: URL path segments
//
// Returns:
//   - int: HTTP response status code
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request, collection string, parts []string) int {
	if len(parts) < 2 {
		http.Error(w, "ID required", http.StatusBadRequest)
		return http.StatusBadRequest
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return http.StatusBadRequest
	}

	idx, found := s.findItemIndex(collection, id)
	if !found {
		http.Error(w, "Item not found", http.StatusNotFound)
		return http.StatusNotFound
	}

	s.data[collection] = append(s.data[collection][:idx], s.data[collection][idx+1:]...)

	if err := s.saveData(); err != nil {
		http.Error(w, "Failed to save data", http.StatusInternalServerError)
		return http.StatusInternalServerError
	}

	w.WriteHeader(http.StatusOK)
	return http.StatusOK
}

// main is the entry point for running the mock JSON server directly.
// It sets up the server with configurable database file and port.
func main() {
	filename := flag.String("db", "db.json", "JSON database file")
	port := flag.Int("port", 3000, "Server port")
	flag.Parse()

	server, err := NewServer(*filename)
	if err != nil {
		log.Fatalf("Error initializing server: %v", err)
	}

	http.HandleFunc("/", server.handleCollection)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("JSON Server running on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
