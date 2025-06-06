package amo

import (
	"net/http"
	"net/url"
)

// SendLead sends a lead request to amoCRM digital pipeline webhook.
// webhookURL should contain full URL to the incoming hook.
func SendLead(webhookURL, name, phone, comment string) error {
	if webhookURL == "" {
		return nil
	}
	data := url.Values{}
	if name != "" {
		data.Set("name", name)
	}
	if phone != "" {
		data.Set("phone", phone)
	}
	if comment != "" {
		data.Set("text", comment)
	}
	resp, err := http.PostForm(webhookURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
