package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/DR-DinoMight/the_surgery/helpers"
	_ "github.com/mattn/go-sqlite3"
)

type Message struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Sender    string `json:"sender"`
	CreatedAt string `json:"created_at"`
}

type Action struct {
	Id     int    `json:"id"`
	Action string `json:"action"`
}

type ChatEventData struct {
	Id   string `json:"id"`
	User struct {
		Id            string    `json:"id"`
		DisplayName   string    `json:"displayName"`
		Color         int       `json:"displayColor"`
		CreatedAt     time.Time `json:"createdAt"`
		PreviousNames []string  `json:"previousNames"`
		IsBot         bool      `json:"isBot"`
		Authenicated  bool      `json:"authenticated"`
	} `json:"user"`
	ClientId  int    `json:"clientId"`
	Body      string `json:"body"`
	RawBody   string `json:"rawBody"`
	Visible   bool   `json:"visible"`
	Timestamp string `json:"timestamp"`
}

type TestEventData struct {
	Id string `json:"id"`
}

type EventType string

const (
	ChatEvent EventType = "CHAT"
	TestEvent EventType = "TEST"
)

// Web hook rout to handle POST requests, and store chat messages in sqlite3 DB
func ChatWebhook(w http.ResponseWriter, r *http.Request) {
	// Get incoming request body
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// get the json body
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Get the Type from the json body
	eventType := EventType(body["type"].(string))

	switch eventType {
	case ChatEvent:
		var eventData ChatEventData
		eventDataBytes, err := json.Marshal(body["eventData"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(eventDataBytes, &eventData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		// check if message doesn't Start with ! if does then it's a command and
		// doesn't want to be stored in DB, or displayed.
		// Instead we should pass it to the commands handler
		if strings.HasPrefix(eventData.Body, "!") {
			ActionCommand(eventData, w, r)
		} else {
			StoreMessage(eventData, w, r)
		}

	}

}

func ActionCommand(eventData ChatEventData, w http.ResponseWriter, r *http.Request) {
	command := strings.TrimPrefix(eventData.Body, "!")
	var message string
	switch command {
	case "help":
		// send help message
		message = fmt.Sprintf("Here are the available commands: !help, !ping, !echo <message>")
		// SendMessage(helpMessage, eventData, w)

	case "ping":
		// send pong message
		message = "Pong!"
		// SendMessage(pongMessage, eventData, w)

	case "echo":
		// echo back the rest of the message
		message = strings.TrimPrefix(eventData.Body, "!echo ")
		// SendMessage(message, eventData, w)

	default:
		// command not recognised
		message = "Unknown command. Type !help to see available commands."
		// SendMessage(unknownCommandMessage, eventData, w)
	}
	log.Println(message)

}

func StoreMessage(eventData ChatEventData, w http.ResponseWriter, r *http.Request) {
	message := eventData.Body
	user := eventData.User.DisplayName
	timestamp := eventData.Timestamp
	colour := eventData.User.Color
	createdAt := time.Now().UTC().Format("2006-01-02T15:04:05.999999999Z")

	// if message, user, or timestamp are empty, return
	if message == "" || user == "" || timestamp == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	message = helpers.ProcessEmojis(message)

	t, _ := time.Parse(time.RFC3339, timestamp)
	//format the timestamp in HH:MM:SS format
	formattedTime := t.Format("15:04")

	//insert message into sqlite3 db
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		panic(err)
	}
	stmt, err := db.Prepare("INSERT INTO messages(message, user, colour, timestamp, created_at) VALUES (?,?,?,?,?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(message, user, colour, formattedTime, createdAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	db.Close()

	fmt.Println(message, user, formattedTime, "inserted into db")
	// return success
	w.WriteHeader(http.StatusOK)
}

func MessageHandler(w http.ResponseWriter, r *http.Request) {
	// Set the response headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	lastTimestamp := time.Now().Format("2006-01-02T15:04:05.999999999Z")

	//if querystring lastTimestamp is provided and not empty string, use that instead

	if last, ok := r.URL.Query()["lastTimestamp"]; ok && len(last) > 1 {
		lastTimestamp = last[0]
		log.Println("lastTime Already Provided, " + lastTimestamp)
	}

	// Create a new context with cancellation support
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Start a goroutine to fetch new messages from the database
	go func() {
		db, err := sql.Open("sqlite3", "database.db")
		if err != nil {
			panic(err)
		}
		defer db.Close()

		for {
			rows, err := db.Query(`SELECT message, user, colour, timestamp, created_at FROM messages WHERE created_at > ? ORDER BY created_at`, lastTimestamp)
			if err != nil {
				panic(err)
			}
			var messages []string
			for rows.Next() {
				var message, user, timestamp, created_at string
				var colour int
				if err := rows.Scan(&message, &user, &colour, &timestamp, &created_at); err != nil {
					panic(err)
				}

				user = helpers.AddColourToUserName(user, colour)
				// message = helpers.ProcessEmojis(message)s
				messageContent := Message{
					Type:      "CHAT",
					Timestamp: timestamp,
					Message:   message,
					Sender:    user,
					CreatedAt: created_at,
				}
				messageJSON, err := json.Marshal(messageContent)
				if err != nil {
					panic(err)
				}

				messages = append(messages, string(messageJSON))

				// messages = append(messages, fmt.Sprintf("%s - %s: %s", timestamp, user, message))
				lastTimestamp = created_at
			}
			rows.Close()

			// Send the new messages as SSE
			for _, msg := range messages {
				select {
				case <-ctx.Done():
					// Client has closed the connection
					return
				default:
					fmt.Println(msg)
					// fmt.Fprintf("event: %s",)
					fmt.Fprintf(w, "event: msg\ndata: %s\n\n", msg)
					w.(http.Flusher).Flush()
				}
			}

			// Wait for a short interval before fetching new messages again
			time.Sleep(1 * time.Second)

			rows, err = db.Query(`SELECT id, action FROM actions where type ='CHAT' and actioned_at is null;`)
			if err != nil {
				panic(err)
			}
			var actions []Action
			for rows.Next() {
				var id int
				var action string
				if err := rows.Scan(&id, &action); err != nil {
					panic(err)
				}

				// push row to actions
				actionObj := Action{
					Id:     id,
					Action: action,
				}
				actions = append(actions, actionObj)
			}
			rows.Close()

			for _, action := range actions {
				select {
				case <-ctx.Done():
					// Client has closed the connection
					return
				default:
					fmt.Fprintf(w, "event: actionMessage\ndata: %s\n\n", action.Action)
					w.(http.Flusher).Flush()

					_, err = db.Exec(fmt.Sprintf("UPDATE actions SET actioned_at = '%s' WHERE id = %d;", time.Now().Format(time.RFC3339), action.Id))

					if err != nil {
						panic(err)
					}
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()

	<-r.Context().Done()
}
