package amo

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"ragbot/internal/config"
	"ragbot/internal/conversation"
	"ragbot/internal/repository"
	"strings"
	"time"
)

const (
	// HTTP request constants
	contentType         = "application/json"
	phoneFieldCode      = "PHONE"
	phoneFieldValueCode = "MOB"
	requestTimeout      = 10 * time.Second

	// API endpoints format
	leadsComplexEndpoint = "https://%s/api/v4/leads/complex"
	contactsEndpoint     = "https://%s/api/v4/contacts"
)

// HTTPClient интерфейс для HTTP-клиента, чтобы можно было заменять его в тестах
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// AmoClient представляет клиента для работы с amoCRM
type AmoClient struct {
	HTTPClient HTTPClient
}

// DefaultAmoClient возвращает клиент с настройками по умолчанию
func DefaultAmoClient() *AmoClient {
	return &AmoClient{
		HTTPClient: http.DefaultClient,
	}
}

// Оставляем глобальный экземпляр для обратной совместимости
var defaultClient = DefaultAmoClient()

// Lead represents an amoCRM lead structure
type lead struct {
	Name               string   `json:"name,omitempty"`
	Embedded           embed    `json:"_embedded,omitempty"`
	CustomFieldsValues []cf     `json:"custom_fields_values,omitempty"`
	Tags               []string `json:"tags,omitempty"`
}

// Value represents a value in custom fields
type value struct {
	Value    string `json:"value,omitempty"`
	EnumCode string `json:"enum_code,omitempty"`
	EnumId   int    `json:"enum_id,omitempty"`
}

// CustomField represents a custom field in amoCRM
type cf struct {
	FieldCode string  `json:"field_code,omitempty"`
	FieldId   int     `json:"field_id,omitempty"`
	Values    []value `json:"values"`
}

// Contact represents a contact in amoCRM
type contact struct {
	Name               string `json:"name,omitempty"`
	FirstName          string `json:"first_name,omitempty"`
	CustomFieldsValues []cf   `json:"custom_fields_values,omitempty"`
	ID                 int    `json:"id,omitempty"`
}

type savedContact struct {
	ID         int    `json:"id"`
	IsDeleted  bool   `json:"is_deleted"`
	IsUnsorted bool   `json:"is_unsorted"`
	RequestID  string `json:"request_id"`
	Links      struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}

// Embed represents embedded data in a lead
type embed struct {
	Contacts []contact `json:"contacts,omitempty"`
}

// ContactResponse представляет ответ от создания контакта в amoCRM
type contactResponse struct {
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
	Embedded struct {
		Contacts []savedContact `json:"contacts"`
	} `json:"_embedded"`
}

// SendLeadToAMO creates a lead in amoCRM using the API v4.
func SendLeadToAMO(repo *repository.Repository, info *conversation.ChatInfo, link string) error {
	return defaultClient.SendLeadToAMO(repo, info, link)
}

// SendLeadToAMO создает лид в amoCRM используя API v4
func (c *AmoClient) SendLeadToAMO(repo *repository.Repository, info *conversation.ChatInfo, link string) error {
	if config.Config.AmoDomain == "" || config.Config.AmoAccessToken == "" {
		log.Println("AMO integration not configured")
		return nil
	}

	loadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	branches := make([]string, 0)
	lowerSummary := strings.ToLower(info.Summary.String)
	for branch, _ := range amoConfig.branchFieldValuesMap {
		lowerBranch := strings.ToLower(branch)
		if strings.Contains(lowerSummary, lowerBranch) {
			branches = append(branches, branch)
		}
	}

	// Generate dynamic tags based on conversation content
	dynamicTags := c.generateDynamicTags(info)

	var cont *savedContact
	var err error
	if info.AmoContactID.Valid {
		cont = &savedContact{ID: int(info.AmoContactID.Int64)}
	} else {
		cont, err = c.createContact(ctx, info.Name.String, info.Phone.String)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to create a contact: %s", err)
			return errors.New(errMsg)
		}
		_ = repo.UpdateAmoContactID(ctx, info.ChatID, sql.NullInt64{Int64: int64(cont.ID), Valid: true})
	}

	// Create lead
	_, err = c.createLead(ctx, cont, info.Title.String, info.Summary.String, info.Interest.String, link, branches, dynamicTags)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create a lead: %s", err)
		return errors.New(errMsg)
	}

	return nil
}

func (c *AmoClient) createContact(ctx context.Context, name, phone string) (*savedContact, error) {
	lead := buildContact(name, phone)
	url := fmt.Sprintf(contactsEndpoint, config.Config.AmoDomain)

	resp, err := c.makeJSONRequest(ctx, url, []any{lead})
	if err != nil || resp == nil || resp.Body == nil {
		errMsg := fmt.Sprintf("Failed to create a contact: %s", err)
		return nil, errors.New(errMsg)
	}
	defer resp.Body.Close()

	var contactResp contactResponse
	if err := json.NewDecoder(resp.Body).Decode(&contactResp); err != nil {
		errMsg := fmt.Sprintf("Failed to parse a contact response: %s", err)
		return nil, errors.New(errMsg)
	} else {
		log.Printf("contact response: %+v", contactResp)
	}

	if len(contactResp.Embedded.Contacts) < 1 {
		return nil, errors.New("No contacts created")
	}

	return &contactResp.Embedded.Contacts[0], nil
}

func (c *AmoClient) createLead(ctx context.Context, cont *savedContact, leadName, summary, interest, link string, branches []string, dynamicTags []string) (*http.Response, error) {
	lead := buildLead(leadName, cont, summary, interest, link, branches, dynamicTags)
	url := fmt.Sprintf(leadsComplexEndpoint, config.Config.AmoDomain)

	return c.makeJSONRequest(ctx, url, []any{lead})
}

func (c *AmoClient) makeJSONRequest(ctx context.Context, url string, payload any) (*http.Response, error) {
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

	resp, err := c.HTTPClient.Do(req)
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

// generateDynamicTags создает теги на основе содержимого разговора
func (c *AmoClient) generateDynamicTags(info *conversation.ChatInfo) []string {
	tags := make([]string, 0)

	// Check if dynamic tags are enabled
	if !amoConfig.dynamicTagsEnabled {
		return tags
	}

	// Add tags based on interest
	if info.Interest.Valid && info.Interest.String != "" {
		tags = append(tags, "Интерес: "+info.Interest.String)
	}

	// Add tags based on summary content
	if info.Summary.Valid && info.Summary.String != "" {
		summary := strings.ToLower(info.Summary.String)
		
		// Add tags based on configured keywords
		for keyword, keywordTags := range amoConfig.keywordTagsMap {
			if strings.Contains(summary, keyword) {
				tags = append(tags, keywordTags...)
			}
		}
	}

	// Add tag based on whether contact already exists
	if info.AmoContactID.Valid {
		tags = append(tags, "Повторный клиент")
	} else {
		tags = append(tags, "Новый клиент")
	}

	// Add tag for RAG bot
	tags = append(tags, "RAG Бот")

	return tags
}

func buildContact(name, phone string) *contact {
	return &contact{
		Name: name,
		CustomFieldsValues: []cf{
			{
				FieldCode: phoneFieldCode,
				Values:    []value{{Value: phone, EnumCode: phoneFieldValueCode}},
			},
		},
	}
}

func buildLead(leadName string, cont *savedContact, summary, interest, link string, branches []string, dynamicTags []string) *lead {
	// Combine static and dynamic tags
	allTags := make([]string, 0, len(amoConfig.tags)+len(dynamicTags))
	allTags = append(allTags, amoConfig.tags...)
	allTags = append(allTags, dynamicTags...)

	l := &lead{
		Name: leadName,
		Embedded: embed{
			Contacts: []contact{
				{
					ID: cont.ID,
				},
			},
		},
		Tags: allTags,
	}

	customFields := []cf{}

	if amoConfig.sourceFieldId != 0 && amoConfig.sourceFieldValueId != 0 {
		customFields = append(customFields, cf{
			FieldId: amoConfig.sourceFieldId,
			Values:  []value{{EnumId: amoConfig.sourceFieldValueId}},
		})
	}

	if amoConfig.summaryFieldId != 0 && summary != "" {
		customFields = append(customFields, cf{
			FieldId: amoConfig.summaryFieldId,
			Values:  []value{{Value: summary}},
		})
	}

	if amoConfig.chatLinkFieldId != 0 && link != "" {
		customFields = append(customFields, cf{
			FieldId: amoConfig.chatLinkFieldId,
			Values:  []value{{Value: link}},
		})
	}

	if amoConfig.interestFieldId != 0 && interest != "" {
		customFields = append(customFields, cf{
			FieldId: amoConfig.interestFieldId,
			Values:  []value{{Value: interest}},
		})
	}

	branchesValues := make([]value, 0)
	for _, branch := range branches {
		if branchValueId, ok := amoConfig.branchFieldValuesMap[branch]; ok {
			branchesValues = append(branchesValues, value{EnumId: branchValueId})
		}
	}

	if len(branchesValues) > 0 {
		l.CustomFieldsValues = append(customFields, cf{
			FieldId: amoConfig.branchFieldId,
			Values:  branchesValues,
		})
	}

	if len(customFields) > 0 {
		l.CustomFieldsValues = customFields
	}

	return l
}
