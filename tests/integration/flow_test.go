//go:build integration
// +build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8080/api"
)

func TestEndToEndFlow(t *testing.T) {
	// This test assumes the API server is running on localhost:8080
	// You must run `docker-compose up` before running this test.

	client := &http.Client{}
	var authToken string
	var userID string

	// 1. Register User
	t.Run("Register", func(t *testing.T) {
		payload := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
			"name":     "Test User",
		}
		body, _ := json.Marshal(payload)
		resp, err := client.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// 2. Login
	t.Run("Login", func(t *testing.T) {
		payload := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		body, _ := json.Marshal(payload)
		resp, err := client.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		
		data := result["data"].(map[string]interface{})
		authToken = data["access_token"].(string)
		
		// Decode user info if available or fetch profile
		// For now assuming token is enough
	})

	var fatherID, childID string

	// 3. Create Father
	t.Run("Create Father", func(t *testing.T) {
		payload := map[string]interface{}{
			"first_name": "Budi",
			"gender":     "MALE",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", baseURL+"/persons", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+authToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		fatherID = result["id"].(string)
	})

	// 4. Create Child
	t.Run("Create Child", func(t *testing.T) {
		payload := map[string]interface{}{
			"first_name": "Ani",
			"gender":     "FEMALE",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", baseURL+"/persons", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+authToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		childID = result["id"].(string)
	})

	// 5. Create Relationship (Father -> Child)
	t.Run("Create Relationship", func(t *testing.T) {
		payload := map[string]interface{}{
			"person_a": childID,
			"person_b": fatherID,
			"type":     "PARENT",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", baseURL+"/relationships", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+authToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// 6. Resolve Relationship
	t.Run("Resolve Relationship", func(t *testing.T) {
		payload := map[string]interface{}{
			"from_person_id": childID,
			"to_person_id":   fatherID,
			"max_depth":      5,
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", baseURL+"/graph/resolve", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+authToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		
		assert.Equal(t, true, result["found"])
		// assert.Equal(t, "FATHER", result["type"]) // Depending on narrator output
	})
}
