package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	initiateUploadId string
	key              string = "test.webm"
)

// executeRequest, creates a new ResponseRecorder
// then executes the request by calling ServeHTTP in the router
// after which the handler writes the response to the response recorder
// which we can then inspect.
func executeRequest(req *http.Request, s *Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	return rr
}

// checkResponseCode is a simple utility to check the response code
// of the response
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestInitiateMultipartUploadHandler(t *testing.T) {
	// Create a New Server Struct
	s := CreateNewServer()
	// Mount Handlers
	s.MountHandlers()

	// Create a New Request
	req, _ := http.NewRequest("GET", fmt.Sprintf("/initiate?key=%s", key), nil)

	// Execute Request
	response := executeRequest(req, s)

	// Check the response code
	checkResponseCode(t, http.StatusAccepted, response.Code)

	// Decode the response body
	var responseBody map[string]interface{}
	err := json.NewDecoder(response.Body).Decode(&responseBody)
	require.NoError(t, err, "Failed to decode response body")

	// Assert the presence and type of uploadId key
	require.NotNil(t, responseBody["uploadId"], "Missing 'uploadId' key in response")

	// Optionally, check the uploadId value if expected
	uploadId, ok := responseBody["uploadId"].(string)
	require.True(t, ok, "uploadId should be a string")

	initiateUploadId = uploadId
}

func TestGetPresignedURLHandler(t *testing.T) {
	// Create a New Server Struct
	s := CreateNewServer()
	// Mount Handlers
	s.MountHandlers()

	// Create a New Request
	req, _ := http.NewRequest("GET", fmt.Sprintf("/presigned?key=%s&partNumber=1&uploadId=%s", key, initiateUploadId), nil)

	// Execute Request
	response := executeRequest(req, s)

	// Check the response code
	checkResponseCode(t, http.StatusCreated, response.Code)

	// Decode the response body
	var responseBody map[string]interface{}
	err := json.NewDecoder(response.Body).Decode(&responseBody)
	require.NoError(t, err, "Failed to decode response body")

	// Assert the presence and type of uploadId key
	require.NotNil(t, responseBody["uploadId"], "Missing 'uploadId' key in response")

	// Optionally, check the uploadId value if expected
	uploadId, ok := responseBody["uploadId"].(string)
	require.True(t, ok, "uploadId should be a string")
	require.Equal(t, initiateUploadId, uploadId, "Incorrect uploadId value")
}
