package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

type Server struct {
	Router *chi.Mux
	// Db, config can be added here
}

var (
	s3Client *s3.Client
)

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		log.Fatalf("Unable to load AWS SDK config, %v", err)
	}
	s3Client = s3.NewFromConfig(cfg)

}

// Handler to initiate the multipart upload
func InitiateMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// Extract bucket and key from environment variables and query parameters
	bucket := os.Getenv("AWS_BUCKET")
	if bucket == "" {
		http.Error(w, "AWS_BUCKET environment variable is not set", http.StatusInternalServerError)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
		return
	}

	// Initiate multipart upload
	output, err := s3Client.CreateMultipartUpload(context.TODO(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		http.Error(w, "Failed to create multipart upload: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send back the UploadId
	response := map[string]interface{}{
		"uploadId": *output.UploadId,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// Handler to generate presigned URL for part upload
func GetPresignedURLHandler(w http.ResponseWriter, r *http.Request) {
	bucket := os.Getenv("AWS_BUCKET")
	key := r.URL.Query().Get("key")
	uploadId := r.URL.Query().Get("uploadId") // Get UploadId from the query parameters
	partNumberStr := r.URL.Query().Get("partNumber")

	if uploadId == "" {
		http.Error(w, "Missing 'uploadId' parameter", http.StatusBadRequest)
		return
	}

	if key == "" {
		http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
		return
	}

	// Convert partNumber string to int32
	partNumber, err := strconv.Atoi(partNumberStr)
	if err != nil || partNumber <= 0 {
		http.Error(w, "Invalid 'partNumber' parameter", http.StatusBadRequest)
		return
	}

	// Generate the presigned URL for this part
	presignClient := s3.NewPresignClient(s3Client)
	expirationDuration := 3600 * time.Second // Set expiration to 1 hour

	req, err := presignClient.PresignUploadPart(context.TODO(), &s3.UploadPartInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(key),
		PartNumber: aws.Int32(int32(partNumber)),
		UploadId:   aws.String(uploadId), // Use the provided UploadId here
	}, s3.WithPresignExpires(expirationDuration))

	if err != nil {
		http.Error(w, "Failed to generate presigned URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate expiration time
	expirationTime := time.Now().Add(expirationDuration).Format(time.RFC3339)

	// Send back the presigned URL and expiration
	response := map[string]interface{}{
		"uploadUrl":  req.URL,
		"uploadId":   uploadId,
		"partNumber": partNumber,
		"expiresAt":  expirationTime,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

func CompleteMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	bucket := os.Getenv("AWS_BUCKET")
	key := r.URL.Query().Get("key")
	uploadId := r.URL.Query().Get("uploadId")

	if key == "" || uploadId == "" {
		http.Error(w, "Missing 'key' or 'uploadId' parameter", http.StatusBadRequest)
		return
	}

	// Parse the parts from the request body
	var requestBody struct {
		Parts []struct {
			PartNumber int    `json:"PartNumber"`
			ETag       string `json:"ETag"`
		} `json:"parts"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Failed to parse request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	var completedParts []types.CompletedPart
	for _, part := range requestBody.Parts {
		completedParts = append(completedParts, types.CompletedPart{
			PartNumber: aws.Int32(int32(part.PartNumber)),
			ETag:       aws.String(part.ETag),
		})
	}

	_, err := s3Client.CompleteMultipartUpload(context.TODO(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadId),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})

	if err != nil {
		http.Error(w, "Failed to complete multipart upload: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Multipart upload completed successfully")
}

func main() {
	s := CreateNewServer()
	s.MountHandlers()

	port := ":8080"
	fmt.Printf("Server is listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(port, s.Router))
}

func CreateNewServer() *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	return s
}

func (s *Server) MountHandlers() {
	r := s.Router
	r.Use(middleware.Logger)
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Mount handler functions with their respective routes
	r.HandleFunc("/initiate", InitiateMultipartUploadHandler)
	r.HandleFunc("/presigned", GetPresignedURLHandler)
	r.HandleFunc("/complete", CompleteMultipartUploadHandler)

}
