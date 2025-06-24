package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDebugVersion2(t *testing.T) {
	config := DefaultAppConfig()
	router := SetupRouter(config)

	// Test V2 product endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/products/1", nil)
	req.Header.Set("API-Version", "2")

	router.ServeHTTP(w, req)

	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
			return
		}

		fmt.Printf("Parsed Response: %+v\n", response)

		if data, exists := response["data"]; exists {
			if product, ok := data.(map[string]interface{}); ok {
				fmt.Printf("Product data: %+v\n", product)

				// Check for V2 specific fields
				if desc, exists := product["description"]; exists {
					fmt.Printf("Description: %v\n", desc)
				} else {
					fmt.Println("Description field missing!")
				}
			}
		}
	} else {
		t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
	}
}
