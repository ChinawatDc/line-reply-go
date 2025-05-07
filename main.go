package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Event struct {
	Source struct {
		Type    string `json:"type"`
		GroupId string `json:"groupId"`
	} `json:"source"`
	ReplyToken string `json:"replyToken"`
}

type WebhookRequest struct {
	Events []Event `json:"events"`
}

func replyMessage(replyToken string, message string) error {
	accessToken := os.Getenv("LINE_API_TOKEN")
	if accessToken == "" {
		return fmt.Errorf("LINE_API_TOKEN not set")
	}

	url := "https://api.line.me/v2/bot/message/reply"
	payload := map[string]interface{}{
		"replyToken": replyToken,
		"messages": []map[string]string{
			{
				"type": "text",
				"text": message,
			},
		},
	}

	body, _ := json.Marshal(payload)
	// fmt.Println(string(body))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("LINE API error: status code %d", resp.StatusCode)
	}

	return nil
}

func doPost(c *gin.Context) {
	var data WebhookRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if len(data.Events) > 0 && data.Events[0].Source.Type == "group" {
		// fmt.Println("Source", data.Events[0].Source)
		groupId := data.Events[0].Source.GroupId
		replyToken := data.Events[0].ReplyToken

		// fmt.Println("replyToken", replyToken)
		// fmt.Println("groupId", groupId)

		message := fmt.Sprintf("This is your Group ID: %s", groupId)

		// fmt.Println("message", message)

		if err := replyMessage(replyToken, message); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found or failed to load")
	}

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "LINE bot is running")
	})

	r.POST("/reply", doPost)
	r.Run(":8080")

}
