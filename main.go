package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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

// ฟังก์ชันสำหรับส่งข้อความตอบกลับไปยัง LINE โดยรับ token เป็นพารามิเตอร์
func replyWithToken(replyToken, message, accessToken string) error {
	if accessToken == "" {
		return fmt.Errorf("access token not set")
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

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("LINE API error: status code %d", resp.StatusCode)
	}

	return nil
}

// ฟังก์ชันหลักสำหรับจัดการ webhook ของ LINE
func handleWebhook(c *gin.Context, accessToken string) {
	var data WebhookRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if len(data.Events) > 0 && data.Events[0].Source.Type == "group" {
		groupId := data.Events[0].Source.GroupId
		replyToken := data.Events[0].ReplyToken
		message := fmt.Sprintf("This is your Group ID: %s", groupId)

		if err := replyWithToken(replyToken, message, accessToken); err != nil {
			log.Printf("Failed to reply: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func main() {
	// โหลดไฟล์ .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or failed to load")
	}

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "LINE bot is running")
	})

	r.POST("/reply", func(c *gin.Context) {
		handleWebhook(c, os.Getenv("LINE_API_TOKEN"))
	})

	r.POST("/reply2", func(c *gin.Context) {
		handleWebhook(c, os.Getenv("LINE_API_TOKEN_2"))
	})

	r.Run(":8080")
}
