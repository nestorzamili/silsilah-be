//go:build integration
// +build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8080/api"
)

func TestGovernanceFlow(t *testing.T) {
	// This test assumes the API server is running on localhost:8080
	// and connects to the same DB as the test runner for role promotion.

	env := SetupTestEnv(t)
	defer env.Teardown()

	client := &http.Client{}

	var requesterToken string
	var reviewerToken string
	var requesterID string
	var changeRequestID string

	// 1. Register Requester
	t.Run("Register Requester", func(t *testing.T) {
		payload := map[string]string{
			"email":    "requester@example.com",
			"password": "password123",
			"name":     "Requester User",
		}
		body, _ := json.Marshal(payload)
		resp, err := client.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// 2. Login Requester
	t.Run("Login Requester", func(t *testing.T) {
		payload := map[string]string{
			"email":    "requester@example.com",
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
		requesterToken = data["access_token"].(string)
		
		// Get User ID from profile or decode token (simulated here by fetching profile)
		req, _ := http.NewRequest("GET", baseURL+"/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+requesterToken)
		respMe, err := client.Do(req)
		require.NoError(t, err)
		defer respMe.Body.Close()
		
		var resultMe map[string]interface{}
		json.NewDecoder(respMe.Body).Decode(&resultMe)
		userData := resultMe["data"].(map[string]interface{})
		requesterID = userData["id"].(string)
	})

	// 3. Register Reviewer
	t.Run("Register Reviewer", func(t *testing.T) {
		payload := map[string]string{
			"email":    "reviewer@example.com",
			"password": "password123",
			"name":     "Reviewer User",
		}
		body, _ := json.Marshal(payload)
		resp, err := client.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(body))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	// 4. Promote Reviewer to EDITOR (via DB)
	t.Run("Promote Reviewer", func(t *testing.T) {
		// We need to find the ID first or just update by email
		_, err := env.DB.Exec("UPDATE users SET role = 'editor' WHERE email = $1", "reviewer@example.com")
		require.NoError(t, err)
	})

	// 5. Login Reviewer
	t.Run("Login Reviewer", func(t *testing.T) {
		payload := map[string]string{
			"email":    "reviewer@example.com",
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
		reviewerToken = data["access_token"].(string)

		req, _ := http.NewRequest("GET", baseURL+"/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+reviewerToken)
		respMe, err := client.Do(req)
		require.NoError(t, err)
		defer respMe.Body.Close()
		
		var resultMe map[string]interface{}
		json.NewDecoder(respMe.Body).Decode(&resultMe)
		userData := resultMe["data"].(map[string]interface{})
		
		// Verify role
		assert.Equal(t, "editor", userData["role"])
	})

	// 6. Requester Submits Change Request (Create Person)
	t.Run("Submit Change Request", func(t *testing.T) {
		payload := map[string]interface{}{
			"entity_type": "person",
			"action":      "create",
			"payload": map[string]interface{}{
				"first_name": "New",
				"last_name":  "Person",
				"gender":     "MALE",
			},
			"requester_note": "Please add this person",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", baseURL+"/change-requests", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+requesterToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].(map[string]interface{})
		changeRequestID = data["id"].(string)
		assert.Equal(t, "pending", data["status"])
		assert.Equal(t, requesterID, data["requested_by"])
	})

	// 7. Reviewer Lists Pending Requests
	t.Run("List Pending Requests", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/change-requests?status=pending", nil)
		req.Header.Set("Authorization", "Bearer "+reviewerToken)
		
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		require.Equal(t, http.StatusOK, resp.StatusCode)
		
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].([]interface{})
		
		found := false
		for _, item := range data {
			cr := item.(map[string]interface{})
			if cr["id"] == changeRequestID {
				found = true
				break
			}
		}
		assert.True(t, found, "Change Request should be in the list")
	})

	// 8. Reviewer Approves Request
	t.Run("Approve Request", func(t *testing.T) {
		payload := map[string]interface{}{
			"action": "approve", // Assuming the endpoint uses an action or status in body
			"note":   "Looks good",
		}
		// Note: The actual endpoint might be POST /change-requests/{id}/approve or PATCH /change-requests/{id}
		// Checking service logic: Approve(ctx, id, reviewerID, note, meta)
		// Usually mapped to POST /api/change-requests/:id/approve or PUT /api/change-requests/:id/status
		// Let's assume standard REST: POST /change-requests/{id}/approve based on service method "Approve" being distinct
		
		// I need to verify the route. I'll assume POST /change-requests/{id}/approve for now.
		// If it fails 404, I'll check routes.
		
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/change-requests/%s/approve", baseURL, changeRequestID), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+reviewerToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 9. Verify Person Created
	t.Run("Verify Person Created", func(t *testing.T) {
		// Wait a bit for async execution if any (though Approve is synchronous in service)
		time.Sleep(100 * time.Millisecond)

		// Search for the person
		req, _ := http.NewRequest("GET", baseURL+"/persons?q=New Person", nil)
		req.Header.Set("Authorization", "Bearer "+requesterToken)
		
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		
		require.Equal(t, http.StatusOK, resp.StatusCode)
		
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		data := result["data"].([]interface{})
		
		found := false
		for _, item := range data {
			p := item.(map[string]interface{})
			if p["first_name"] == "New" && p["last_name"] == "Person" {
				found = true
				break
			}
		}
		assert.True(t, found, "Person should be created and visible")
	})
}
