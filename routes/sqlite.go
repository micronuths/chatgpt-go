package routes

import (
	"chatgpt-go/global"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"chatgpt-go/pkg/lemur"

	_ "modernc.org/sqlite"
)

type Chat struct {
	Id        int                           `json:"id"`
	MessageId string                        `json:"message_id"`
	Messages  []lemur.ChatCompletionMessage `json:"messages"`
	Message   lemur.ChatCompletionMessage   `json:"message"`
}

type ChatStorage struct {
	db *sql.DB
}

func NewChatStorage(path ...string) (*ChatStorage, error) {

	dbpath := global.Config.System.DatabasePath
	if dbpath == "" {
		cwd, _ := os.Getwd()
		dbpath = filepath.Join(cwd, "database.sqlite")
	}
	if len(path) != 0 {
		dbpath = path[len(path)-1]
	}
	db, err := sql.Open("sqlite", dbpath)

	if err != nil {
		log.Fatal(err)
	}

	// Create table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS chat (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
			message_id varchar(255),
            messages TEXT,
			parent_message_id varchar(255)
        );
    `)
	if err != nil {
		log.Fatal(err)
	}
	return &ChatStorage{db: db}, nil
}

func (c *ChatStorage) GetContextMessages(messageID string) ([]lemur.ChatCompletionMessage, error) {
	var chatCompletionMessageList = make([]lemur.ChatCompletionMessage, 0)

	for {
		chatCompletionMessage, parentID, err := c.GetMessage(messageID)
		if err != nil {
			fmt.Printf("Error when c.GetMessage(messageID=%s),%s", messageID, err)
			break
		}
		chatCompletionMessageList = append(chatCompletionMessageList, chatCompletionMessage)

		// fmt.Println(messageID, "->", parentID)
		messageID = parentID
		if messageID == "chatcmpl-start" {
			break
		}

	}
	return chatCompletionMessageList, nil
}

// 目标根据messageID 返回对应结构体
func (c *ChatStorage) GetMessage(messageID string) (lemur.ChatCompletionMessage, string, error) {
	var messagesStr string
	var parentMessageIDStr string
	err := c.db.QueryRow("SELECT messages FROM chat WHERE message_id = ?", messageID).Scan(&messagesStr)
	if err != nil {
		return lemur.ChatCompletionMessage{}, "", err
	}
	err = c.db.QueryRow("SELECT parent_message_id FROM chat WHERE message_id = ?", messageID).Scan(&parentMessageIDStr)
	if err != nil {
		return lemur.ChatCompletionMessage{}, "", err
	}

	var chat Chat
	err = json.Unmarshal([]byte(messagesStr), &chat.Message)
	if err != nil {
		return lemur.ChatCompletionMessage{}, "", err
	}

	return chat.Message, parentMessageIDStr, nil
}

// 原始函数-返回列表
func (c *ChatStorage) GetMessages(messageID string) ([]lemur.ChatCompletionMessage, error) {
	var messagesStr string

	err := c.db.QueryRow("SELECT messages FROM chat WHERE message_id = ?", messageID).Scan(&messagesStr)
	if err != nil {
		return nil, err
	}

	var chat Chat
	err = json.Unmarshal([]byte(messagesStr), &chat.Messages)
	if err != nil {
		return nil, err
	}

	return chat.Messages, nil
}

/*
添加记录
*/
func (c *ChatStorage) AddMessage(currentMessageId string, parentMessageId string, message lemur.ChatCompletionMessage) error {

	updatedMessages, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Marshal error: %v\n", err)
		return err
	}

	_, err = c.db.Exec("INSERT INTO chat (message_id,messages,parent_message_id) VALUES (?,?,?)", currentMessageId, string(updatedMessages), parentMessageId)
	if err != nil {
		fmt.Printf("UPDATE chat error: %v\n", err)
		return err
	}

	return nil
}

func (c *ChatStorage) Close() {
	c.db.Close()
}
