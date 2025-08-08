package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var events = []map[string]interface{}{
	{
		"specversion": "1.0",
		"type":        "grant.created",
		"source":      "/nylas/system",
		"id":          "mock-id",
		"time":        1234567890,
		"data": map[string]interface{}{
			"application_id": "NYLAS_APPLICATION_ID",
			"object": map[string]interface{}{
				"code":           25012,
				"grant_id":       "NYLAS_GRANT_ID",
				"integration_id": "NYLAS_INTEGRATION_ID",
				"login_id":       "mock-login-id",
				"provider":       "google",
			},
		},
	},
	{
		"specversion": "1.0",
		"type":        "grant.updated",
		"source":      "/nylas/system",
		"id":          "mock-id",
		"time":        123456789,
		"data": map[string]interface{}{
			"application_id": "NYLAS_APPLICATION_ID",
			"object": map[string]interface{}{
				"code":                  25014,
				"grant_id":              "NYLAS_GRANT_ID",
				"integration_id":        "NYLAS_INTEGRATION_ID",
				"provider":              "microsoft",
				"reauthentication_flag": false,
			},
		},
	},
	{
		"specversion": "1.0",
		"type":        "grant.deleted",
		"source":      "/nylas/system",
		"id":          "mock-id",
		"time":        1234567890,
		"data": map[string]interface{}{
			"application_id": "NYLAS_APPLICATION_ID",
			"object": map[string]interface{}{
				"code":           25013,
				"grant_id":       "NYLAS_GRANT_ID",
				"integration_id": "NYLAS_INTEGRATION_ID",
				"provider":       "google",
			},
		},
	},
	{
		"specversion":              "1.0",
		"type":                     "message.created",
		"source":                   "/google/emails/realtime",
		"id":                       "<WEBHOOK_ID>",
		"time":                     1723821985,
		"webhook_delivery_attempt": 1,
		"data": map[string]interface{}{
			"application_id": "<NYLAS_APPLICATION_ID>",
			"object": map[string]interface{}{
				"attachments": map[string]interface{}{
					"content_disposition": "attachment; filename=\"image.jpg\"",
					"content_id":          "<CID>",
					"content_type":        "image/jpeg; name=\"image.jpg\"",
					"filename":            "image.jpg",
					"grant_id":            "<NYLAS_GRANT_ID>",
					"id":                  "<ATTACHMENT_ID>",
					"is_inline":           false,
					"size":                4860136,
				},
				"bcc": map[string]interface{}{
					"email": "leyah@example.com",
				},
				"body": "<div dir=\"ltr\">Test with attachments</div>\r\n",
				"cc": map[string]interface{}{
					"email": "kaveh@example.com",
				},
				"date":    1723821981,
				"folders": []string{"SENT"},
				"from": map[string]interface{}{
					"email": "swag@example.com",
					"name":  "Nylas Swag",
				},
				"grant_id": "<NYLAS_GRANT_ID>",
				"id":       "<MESSAGE_ID>",
				"metadata": map[string]interface{}{
					"key1": "all-meetings",
					"key2": "on-site",
				},
				"object":    "message",
				"reply_to":  map[string]interface{}{},
				"snippet":   "This message has an attachment. yippee!",
				"starred":   false,
				"subject":   "Let's send an attachment",
				"thread_id": "<THREAD_ID>",
				"to": map[string]interface{}{
					"email": "nyla@example.com",
				},
				"unread": false,
			},
		},
	},
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
	// Set the headers so that the client knows what data they're receiving
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Iterate over mock events
	for _, payload := range events {
		// Convert the map into a compact JSON string so that it can be sent over a network
		jsonBytesCompact, err := json.Marshal(payload)
		if err != nil { // Couldn't JSONify one of the events, throw error
			log.Fatal(err)
		}
		// Print the JSON string to the buffer
		fmt.Fprintf(w, "data: %v\n", string(jsonBytesCompact))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush() // Send the current buffer as a chunk
		}
		// Simulate waiting for an event
		time.Sleep(1 * time.Second)
	}
	// End connection
	fmt.Fprint(w, ":Stream finished.\n")
}

func main() {
	http.HandleFunc("/stream", streamHandler)
	fmt.Println("Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}
