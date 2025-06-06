package amo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"ragbot/internal/config"
	"time"
)

const (
	// HTTP request constants
	contentType    = "application/json"
	noteTypeCommon = "common"
	phoneFieldCode = "PHONE"
	phoneEnumCode  = "WORK"
	requestTimeout = 10 * time.Second

	// API endpoints format
	leadsComplexEndpoint = "https://%s/api/v4/leads/complex"
	leadsNotesEndpoint   = "https://%s/api/v4/leads/%d/notes"
)

// Lead represents an amoCRM lead structure
type lead struct {
	Name     string `json:"name,omitempty"`
	Embedded embed  `json:"_embedded,omitempty"`
}

// Value represents a value in custom fields
type value struct {
	Value    string `json:"value"`
	EnumCode string `json:"enum_code"`
}

// CustomField represents a custom field in amoCRM
type cf struct {
	FieldCode string  `json:"field_code"`
	Values    []value `json:"values"`
}

// Contact represents a contact in amoCRM
type contact struct {
	FirstName          string `json:"first_name,omitempty"`
	CustomFieldsValues []cf   `json:"custom_fields_values,omitempty"`
}

// Embed represents embedded data in a lead
type embed struct {
	Contacts []contact `json:"contacts,omitempty"`
}

// LeadResponse represents the response from creating a lead
type leadResponse struct {
	Embedded struct {
		Leads []struct {
			ID int `json:"id"`
		} `json:"leads"`
	} `json:"_embedded"`
}

// SendLead creates a lead in amoCRM using the API v4.
func SendLead(name, phone, comment string) error {
	if config.Config.AmoDomain == "" || config.Config.AmoAccessToken == "" {
		log.Println("AMO integration not configured")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	// Create lead
	resp, err := createLead(ctx, name, phone)
	if err != nil {
		return fmt.Errorf("failed to create lead: %w", err)
	}
	defer resp.Body.Close()

	var lr leadResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return fmt.Errorf("failed to decode lead response: %w", err)
	}

	// If there's no comment or no leads were created, we're done
	if comment == "" || len(lr.Embedded.Leads) == 0 {
		return nil
	}

	// Create note
	if _, err = createNote(ctx, lr, comment); err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	return nil
}

func createLead(ctx context.Context, name, phone string) (*http.Response, error) {
	lead := buildLead(name, phone)
	url := fmt.Sprintf(leadsComplexEndpoint, config.Config.AmoDomain)

	return makeJSONRequest(ctx, url, []any{lead})
}

func createNote(ctx context.Context, lr leadResponse, comment string) (*http.Response, error) {
	noteURL := fmt.Sprintf(leadsNotesEndpoint, config.Config.AmoDomain, lr.Embedded.Leads[0].ID)
	noteBody := []map[string]any{
		{
			"note_type": noteTypeCommon,
			"params":    map[string]string{"text": comment},
		},
	}

	return makeJSONRequest(ctx, noteURL, noteBody)
}

func makeJSONRequest(ctx context.Context, url string, payload any) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.Config.AmoAccessToken)
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("amoCRM API error: %s: %s", resp.Status, string(data))
	}

	return resp, nil
}

func buildLead(name, phone string) *lead {
	return &lead{
		Name: name,
		Embedded: embed{
			Contacts: []contact{
				{
					FirstName: name,
					CustomFieldsValues: []cf{
						{
							FieldCode: phoneFieldCode,
							Values:    []value{{Value: phone, EnumCode: phoneEnumCode}},
						},
					},
				},
			},
		},
	}
}
