package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// TS3Client is the main object
type TS3Client struct {
	connection *net.TCPConn
}

// Client is a user who uses TS
type Client struct {
	lastClientID string
	nickname     string
}

// Event is a minimal info that we need
type Event struct {
	action         string
	clientID       string
	clientUniqueID string
	nickname       string
	receivedAt     time.Time
}

// Interface for mocking time in the tests
type Clock interface {
	Now() time.Time
}

// Implement clock for getting real time using "time" library
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// Exec sends the command to the TS3 server
func (client *TS3Client) Exec(command string) {
	log.Printf("Send: %s\n", command)
	command = fmt.Sprintf("%s\n", command)
	_, err := client.connection.Write([]byte(command))
	if err != nil {
		log.Fatalf("Can't send the commad: %s\n%v", command, err)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	clients := make(map[string]Client)
	connectionsChannel := make(chan *TS3Client)

	go func() {
		for ts3client := range connectionsChannel {
			defer Close(ts3client.connection)
		}
	}()

	connectionsChannel <- connect(&clients, connectionsChannel)

	select {}
}

func connect(clients *map[string]Client, connectionsChannel chan *TS3Client) *TS3Client {
	client := new(TS3Client)

	var err error
	ts3Adress := fmt.Sprintf("%s:%s", os.Getenv("TS3_DISCORD_BOT_HOST"), os.Getenv("TS3_DISCORD_BOT_PORT"))
	tcpAddress, err := net.ResolveTCPAddr("tcp4", ts3Adress)
	if err != nil {
		log.Fatalf("Can't figure out the address: %v\n%s", err, ts3Adress)
	}

	client.connection, err = net.DialTCP("tcp", nil, tcpAddress)
	if err != nil {
		log.Fatalf("Can't connect to the TS3 server: %v", err)
	}

	if err = client.connection.SetKeepAlive(true); err != nil {
		log.Fatalf("Can't set keep alive connection: %v", err)
	}

	scanner := bufio.NewScanner(client.connection)
	scanner.Split(scanTS3Lines)

	responseChannel := make(chan string)

	go func() {
		for scanner.Scan() {
			responseChannel <- scanner.Text()
		}
		log.Printf("Connection closed! Error: %v", scanner.Err())
		connectionsChannel <- connect(clients, connectionsChannel)
	}()

	go func() {
		for response := range responseChannel {
			if strings.Index(response, "notify") == 0 {
				log.Printf("Notification: %v\n", response)
				event := parseNotification(response, realClock{})
				populateNickname(clients, &event)
				sendEventToDiscord(event)
			} else {
				log.Printf("Response: %v\n", response)
			}
		}
	}()

	log.Println("Try to login")
	command := "login client_login_name=" + os.Getenv("TS3_DISCORD_BOT_LOGIN") + " client_login_password=" + os.Getenv("TS3_DISCORD_BOT_PASSWORD")
	client.Exec(command)

	log.Println("Use the first server")
	command = "use sid=1"
	client.Exec(command)

	log.Println("Register a notifier")
	command = "servernotifyregister event=server"
	client.Exec(command)

	return client
}

func populateNickname(clients *map[string]Client, event *Event) {
	if event.action == "has connected" {
		(*clients)[event.clientID] = Client{
			lastClientID: event.clientID,
			nickname:     event.nickname,
		}
	} else if event.nickname == "" {
		event.nickname = (*clients)[event.clientID].nickname
	}
}

func sendEventToDiscord(event Event) {
	if event.nickname == "" {
		event.nickname = "Anonymous"
	}

	url := os.Getenv("TS3_DISCORD_BOT_WEBHOOK_URL")
	loc, err := time.LoadLocation(os.Getenv("TS3_DISCORD_BOT_TIMEZONE"))
	if err != nil {
		log.Fatalf("Failed to load the Timezone: %v\n", err)
	}
	timestamp := event.receivedAt.In(loc)
	message := fmt.Sprintf("%02d:%02d:%02d %s %s", timestamp.Hour(), timestamp.Minute(), timestamp.Second(), event.nickname, event.action)
	jsonString := fmt.Sprintf(`{"content":"%s"}`, message)
	json := []byte(jsonString)

	log.Printf("Send to Discord: %s\n", jsonString)

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	if err != nil {
		log.Fatalf("Failed to create a request: %v\n", err)
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		log.Printf("Failed to send an event to Discord: %v\n", err)
	}
	defer Close(response.Body)

	log.Printf("Request status: %v\n", response.Status)
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("An error occured during parsing the response: %v\n", err)
	}
	log.Printf("Request body: %v\n", string(body))
}

func parseNotification(notification string, clock Clock) Event {
	sep := regexp.MustCompile("(\\w+)=(\\S+)")
	splitted := sep.FindAllStringSubmatch(notification, -1)
	var event = Event{receivedAt: clock.Now().UTC()}
	for _, parameter := range splitted {
		key := parameter[1]
		value := parameter[2]

		switch key {
		case "clid":
			event.clientID = value
		case "client_nickname":
			event.nickname = strings.Replace(value, "\\s", " ", -1)
		case "client_unique_identifier":
			event.clientUniqueID = value
		case "reasonid":
			switch value {
			case "0":
				event.action = "has connected"
			case "3":
				event.action = "has lost the connection"
			case "8":
				event.action = "has disconnected"
			}
		}
	}
	return event
}

// Close is a helper function for deferring closing IO
func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatalf("%v\n%v", c, err)
	}
}

// This function is almost exactly like bufio.ScanLines except the \r\n are in opposite positions
func scanTS3Lines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte("\n\r")); i >= 0 {
		return i + 2, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
