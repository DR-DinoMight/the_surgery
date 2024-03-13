package helpers

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/eiannone/keyboard"
)

func CtrlK() {
	go func() {
		// Listen for a key press of 'ctrl-k' to truncate table
		fmt.Println("Press ctrl-k to reinitialise table")

		// Initialize the keyboard
		err := keyboard.Open()
		if err != nil {
			fmt.Println("Error initializing keyboard:", err)
			os.Exit(1)
		}
		defer keyboard.Close()

		for {
			_, key, err := keyboard.GetKey()
			if err != nil {
				fmt.Println("Error reading keyboard input:", err)
				continue
			}

			if key == keyboard.KeyCtrlK {
				// Truncate the table
				//check if sqllite3 db exists, if not create it
				db, err := sql.Open("sqlite3", "database.db")
				if err != nil {
					panic(err)
				}

				// truncate the table
				_, err = db.Exec("DELETE FROM messages")
				if err != nil {
					panic(err)
				}

				//insert an action into the action table
				_, err = db.Exec(fmt.Sprintf("INSERT INTO actions(action, type, created_at) VALUES ('%s','%s', '%s')",
					"clear",
					"CHAT",
					time.Now().Format("2006-01-02T15:04:05.999999999Z")))
				if err != nil {
					panic(err)
				}

				db.Close()
				if err != nil {
					fmt.Println("Error truncating table:", err)
				} else {
					fmt.Println("Table truncated successfully!")
				}
				break
			} else if key == keyboard.KeyCtrlJ {
				OpenConfigFile()
			}
		}
	}()
}
