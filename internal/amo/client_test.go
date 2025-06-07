package amo

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"ragbot/internal/config"
	"strings"
	"testing"
)

// MockHTTPClient мокирует HTTP-клиент для тестов
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestSendLead(t *testing.T) {
	// Убедимся, что конфигурация инициализирована
	if config.Config == nil {
		config.Config = &config.AppConfig{}
	}

	// Сохраняем настоящие значения конфигурации и восстановим после теста
	origAmoDomain := config.Config.AmoDomain
	origAmoAccessToken := config.Config.AmoAccessToken
	defer func() {
		config.Config.AmoDomain = origAmoDomain
		config.Config.AmoAccessToken = origAmoAccessToken
	}()

	// Тестовые данные
	testCases := []struct {
		name           string
		setupConfig    func()
		leadResp       interface{}
		noteResp       interface{}
		leadStatusCode int
		noteStatusCode int
		inputName      string
		inputPhone     string
		inputComment   string
		wantErr        bool
	}{
		{
			name: "successful_lead_and_note_creation",
			setupConfig: func() {
				config.Config.AmoDomain = "test.amocrm.ru"
				config.Config.AmoAccessToken = "test_token"
			},
			leadResp: leadResponse{
				Embedded: struct {
					Leads []struct {
						ID int "json:\"id\""
					} "json:\"leads\""
				}{
					Leads: []struct {
						ID int "json:\"id\""
					}{
						{ID: 12345},
					},
				},
			},
			noteResp:       map[string]interface{}{"success": true},
			leadStatusCode: http.StatusOK,
			noteStatusCode: http.StatusOK,
			inputName:      "Тестовый Клиент",
			inputPhone:     "+7123456789",
			inputComment:   "Тестовый комментарий",
			wantErr:        false,
		},
		{
			name: "missing_amo_config",
			setupConfig: func() {
				config.Config.AmoDomain = ""
				config.Config.AmoAccessToken = ""
			},
			leadResp:       nil,
			noteResp:       nil,
			leadStatusCode: http.StatusOK,
			noteStatusCode: http.StatusOK,
			inputName:      "Тестовый Клиент",
			inputPhone:     "+7123456789",
			inputComment:   "Тестовый комментарий",
			wantErr:        false,
		},
		{
			name: "lead_creation_failure",
			setupConfig: func() {
				config.Config.AmoDomain = "test.amocrm.ru"
				config.Config.AmoAccessToken = "test_token"
			},
			leadResp:       map[string]interface{}{"error": "Unauthorized"},
			leadStatusCode: http.StatusUnauthorized,
			noteResp:       nil,
			noteStatusCode: http.StatusOK,
			inputName:      "Тестовый Клиент",
			inputPhone:     "+7123456789",
			inputComment:   "Тестовый комментарий",
			wantErr:        true,
		},
		{
			name: "lead_created_note_failure",
			setupConfig: func() {
				config.Config.AmoDomain = "test.amocrm.ru"
				config.Config.AmoAccessToken = "test_token"
			},
			leadResp: leadResponse{
				Embedded: struct {
					Leads []struct {
						ID int "json:\"id\""
					} "json:\"leads\""
				}{
					Leads: []struct {
						ID int "json:\"id\""
					}{
						{ID: 12345},
					},
				},
			},
			noteResp:       map[string]interface{}{"error": "Unauthorized"},
			leadStatusCode: http.StatusOK,
			noteStatusCode: http.StatusUnauthorized,
			inputName:      "Тестовый Клиент",
			inputPhone:     "+7123456789",
			inputComment:   "Тестовый комментарий",
			wantErr:        true,
		},
		{
			name: "successful_lead_no_comment",
			setupConfig: func() {
				config.Config.AmoDomain = "test.amocrm.ru"
				config.Config.AmoAccessToken = "test_token"
			},
			leadResp: leadResponse{
				Embedded: struct {
					Leads []struct {
						ID int "json:\"id\""
					} "json:\"leads\""
				}{
					Leads: []struct {
						ID int "json:\"id\""
					}{
						{ID: 12345},
					},
				},
			},
			noteResp:       nil,
			leadStatusCode: http.StatusOK,
			noteStatusCode: http.StatusOK,
			inputName:      "Тестовый Клиент",
			inputPhone:     "+7123456789",
			inputComment:   "", // Пустой комментарий
			wantErr:        false,
		},
		{
			name: "malformed_response",
			setupConfig: func() {
				config.Config.AmoDomain = "test.amocrm.ru"
				config.Config.AmoAccessToken = "test_token"
			},
			leadResp:       "{malformed json", // Некорректный JSON
			leadStatusCode: http.StatusOK,
			noteResp:       nil,
			noteStatusCode: http.StatusOK,
			inputName:      "Тестовый Клиент",
			inputPhone:     "+7123456789",
			inputComment:   "Тестовый комментарий",
			wantErr:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Устанавливаем конфигурацию для этого теста
			tc.setupConfig()

			if config.Config.AmoDomain == "" || config.Config.AmoAccessToken == "" {
				// Если конфигурация AMO отсутствует, просто вызываем функцию напрямую
				err := SendLead(tc.inputName, tc.inputPhone, tc.inputComment)

				if err != nil && !tc.wantErr {
					t.Errorf("SendLead() error = %v, wantErr %v", err, tc.wantErr)
				}
				return
			}

			// Счетчик запросов для определения, какой ответ вернуть
			requestCount := 0

			// Создаем мок HTTP-клиента
			mockClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// Проверяем заголовки
					if req.Header.Get("Authorization") != "Bearer "+config.Config.AmoAccessToken {
						return &http.Response{
							StatusCode: http.StatusUnauthorized,
							Body:       io.NopCloser(strings.NewReader(`{"error": "Unauthorized"}`)),
						}, nil
					}

					// Определяем, какой ответ вернуть в зависимости от порядка запроса
					var respBody interface{}
					var statusCode int

					if requestCount == 0 {
						// Это запрос для создания лида
						respBody = tc.leadResp
						statusCode = tc.leadStatusCode
					} else {
						// Это запрос для создания заметки
						respBody = tc.noteResp
						statusCode = tc.noteStatusCode
					}

					requestCount++

					// Создаем тело ответа
					var respBodyReader io.ReadCloser

					if respBody != nil {
						if s, ok := respBody.(string); ok && s == "{malformed json" {
							// Для теста с некорректным JSON
							respBodyReader = io.NopCloser(strings.NewReader(s))
						} else {
							// Для корректных JSON-ответов
							jsonData, _ := json.Marshal(respBody)
							respBodyReader = io.NopCloser(bytes.NewReader(jsonData))
						}
					} else {
						respBodyReader = io.NopCloser(strings.NewReader("{}"))
					}

					return &http.Response{
						StatusCode: statusCode,
						Body:       respBodyReader,
					}, nil
				},
			}

			// Создаем клиент AMO с моком HTTP-клиента
			client := &AmoClient{
				HTTPClient: mockClient,
			}

			// Вызываем тестируемую функцию
			err := client.SendLead(tc.inputName, tc.inputPhone, tc.inputComment)

			// Проверяем результат
			if (err != nil) != tc.wantErr {
				t.Errorf("SendLead() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestBuildLead(t *testing.T) {
	name := "Тестовый Клиент"
	phone := "+7123456789"

	lead := buildLead(name, phone)

	// Проверяем структуру лида
	if lead.Name != name {
		t.Errorf("buildLead() name = %v, want %v", lead.Name, name)
	}

	if len(lead.Embedded.Contacts) != 1 {
		t.Fatalf("buildLead() contacts count = %v, want 1", len(lead.Embedded.Contacts))
	}

	contact := lead.Embedded.Contacts[0]
	if contact.FirstName != name {
		t.Errorf("buildLead() contact name = %v, want %v", contact.FirstName, name)
	}

	if len(contact.CustomFieldsValues) != 1 {
		t.Fatalf("buildLead() custom fields count = %v, want 1", len(contact.CustomFieldsValues))
	}

	customField := contact.CustomFieldsValues[0]
	if customField.FieldCode != phoneFieldCode {
		t.Errorf("buildLead() field code = %v, want %v", customField.FieldCode, phoneFieldCode)
	}

	if len(customField.Values) != 1 {
		t.Fatalf("buildLead() values count = %v, want 1", len(customField.Values))
	}

	if customField.Values[0].Value != phone {
		t.Errorf("buildLead() phone value = %v, want %v", customField.Values[0].Value, phone)
	}

	if customField.Values[0].EnumCode != phoneEnumCode {
		t.Errorf("buildLead() enum code = %v, want %v", customField.Values[0].EnumCode, phoneEnumCode)
	}
}

// TestMakeJSONRequest проверяет функцию makeJSONRequest
func TestMakeJSONRequest(t *testing.T) {
	// Сохраняем и восстанавливаем оригинальный токен
	origToken := config.Config.AmoAccessToken
	defer func() {
		config.Config.AmoAccessToken = origToken
	}()
	config.Config.AmoAccessToken = "test_token"

	// Создаем мок HTTP-клиента
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Проверяем заголовки
			if req.Header.Get("Authorization") != "Bearer test_token" {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(strings.NewReader(`{"error": "Unauthorized"}`)),
				}, nil
			}

			if req.Header.Get("Content-Type") != contentType {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(strings.NewReader(`{"error": "Invalid content type"}`)),
				}, nil
			}

			// Проверяем метод
			if req.Method != http.MethodPost {
				return &http.Response{
					StatusCode: http.StatusMethodNotAllowed,
					Body:       io.NopCloser(strings.NewReader(`{"error": "Method not allowed"}`)),
				}, nil
			}

			// Декодируем тело запроса для проверки
			var body map[string]interface{}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(strings.NewReader(`{"error": "Bad request body"}`)),
				}, nil
			}

			// Проверяем, что тело содержит ожидаемые поля
			if _, ok := body["test"]; !ok {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(strings.NewReader(`{"error": "Missing test field"}`)),
				}, nil
			}

			// Отправляем успешный ответ
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"result": "success"}`)),
			}, nil
		},
	}

	// Создаем клиент AMO с моком HTTP-клиента
	client := &AmoClient{
		HTTPClient: mockClient,
	}

	// Вызываем тестируемую функцию
	ctx := context.Background()
	resp, err := client.makeJSONRequest(ctx, "https://example.com", map[string]string{"test": "data"})

	// Проверяем результат
	if err != nil {
		t.Errorf("makeJSONRequest() unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("makeJSONRequest() status code = %v, want %v", resp.StatusCode, http.StatusOK)
	}

	// Мок для тестирования ошибки
	mockErrorClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`{"error": "Test error"}`)),
			}, nil
		},
	}

	// Создаем клиент AMO с моком HTTP-клиента
	errorClient := &AmoClient{
		HTTPClient: mockErrorClient,
	}

	// Проверяем обработку ошибки
	_, err = errorClient.makeJSONRequest(ctx, "https://example.com", map[string]string{"test": "data"})
	if err == nil {
		t.Error("makeJSONRequest() with error response should return error")
	}
}
