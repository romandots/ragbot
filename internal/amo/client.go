package amo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SendLead creates a lead in amoCRM using the API v4.
// domain should be like "example.amocrm.ru" without protocol.
// accessToken is the OAuth2 access token for the account.
func SendLead(domain, accessToken, name, phone, comment string) error {
	if domain == "" || accessToken == "" {
		return nil
	}

	// Build request body for POST /api/v4/leads/complex
	type value struct {
		Value    string `json:"value"`
		EnumCode string `json:"enum_code"`
	}
	type cf struct {
		FieldCode string  `json:"field_code"`
		Values    []value `json:"values"`
	}
	type contact struct {
		FirstName          string `json:"first_name,omitempty"`
		CustomFieldsValues []cf   `json:"custom_fields_values,omitempty"`
	}
	type embed struct {
		Contacts []contact `json:"contacts,omitempty"`
	}
	lead := struct {
		Name     string `json:"name,omitempty"`
		Embedded embed  `json:"_embedded,omitempty"`
	}{
		Name: name,
		Embedded: embed{
			Contacts: []contact{
				{
					FirstName: name,
					CustomFieldsValues: []cf{
						{
							FieldCode: "PHONE",
							Values:    []value{{Value: phone, EnumCode: "WORK"}},
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal([]any{lead})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://%s/api/v4/leads/complex", domain)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("amoCRM lead create failed: %s: %s", resp.Status, string(data))
	}

	var lr struct {
		Embedded struct {
			Leads []struct {
				ID int `json:"id"`
			} `json:"leads"`
		} `json:"_embedded"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return err
	}
	if comment == "" || len(lr.Embedded.Leads) == 0 {
		return nil
	}

	// Add a note with the comment
	noteURL := fmt.Sprintf("https://%s/api/v4/leads/%d/notes", domain, lr.Embedded.Leads[0].ID)
	noteBody := []map[string]any{
		{
			"note_type": "common",
			"params":    map[string]string{"text": comment},
		},
	}
	nb, err := json.Marshal(noteBody)
	if err != nil {
		return err
	}
	req, err = http.NewRequest("POST", noteURL, bytes.NewReader(nb))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()
	if resp2.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(resp2.Body)
		return fmt.Errorf("amoCRM add note failed: %s: %s", resp2.Status, string(data))
	}
	return nil
}
