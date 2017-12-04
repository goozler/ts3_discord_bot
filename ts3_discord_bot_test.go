package main

import (
	"fmt"
	"testing"
	"time"
)

type NotificationParams struct {
	clid                     string
	client_nickname          string
	client_unique_identifier string
	reasonid                 string
}

var freezedTime, _ = time.Parse(time.RFC3339, "2017-11-25T22:35:05+04:00")
var freezedTimeUTC = freezedTime.UTC()

// Implement clock to return static time
type testClock struct{}

func (testClock) Now() time.Time {
	return freezedTime
}

var actionTests = []struct {
	reasonid string // input
	expected string // expected result
}{
	{"0", "has connected"},
	{"3", "has lost the connection"},
	{"8", "has disconnected"},
}

func TestParseNotification(t *testing.T) {
	expectedClientID := "2"
	expectedNickname := "nickname with space"
	expectedClientUniqueID := "TgZnzr89N=KVNbyQtoy9V3+ZPzk="

	for _, tAction := range actionTests {
		notification := buildNotification(NotificationParams{
			clid:                     expectedClientID,
			client_nickname:          "nickname\\swith\\sspace",
			client_unique_identifier: expectedClientUniqueID,
			reasonid:                 tAction.reasonid,
		})

		event := parseNotification(notification, testClock{})

		if event.clientID != expectedClientID {
			t.Fatalf("Expected clientID %s but got %s", expectedClientID, event.clientID)
		}
		if event.nickname != expectedNickname {
			t.Fatalf("Expected nickname %s but got %s", expectedNickname, event.nickname)
		}
		if event.clientUniqueID != expectedClientUniqueID {
			t.Fatalf("Expected clientUniqueID %s but got %s", expectedClientUniqueID, event.clientUniqueID)
		}
		if event.action != tAction.expected {
			t.Fatalf("Expected action %s but got %s", tAction.expected, event.action)
		}
		if !event.receivedAt.Equal(freezedTimeUTC) {
			t.Fatalf("Expected time %s but got %s", freezedTimeUTC, event.receivedAt)
		}
	}
}

func TestPopulateNickname(t *testing.T) {
	clientID := "someID"
	nickname := "testNickname"

	clients := make(map[string]Client)

	var event = Event{
		clientID: clientID,
		nickname: nickname,
		action:   "has connected",
	}

	populateNickname(&clients, &event)

	if client, ok := clients[clientID]; ok {
		if client.nickname != nickname {
			t.Fatal("Client's nickname is not saved")
		}
		if client.lastClientID != clientID {
			t.Fatal("Client's ID is not saved")
		}
	} else {
		t.Fatal("Client is not saved")
	}

	event = Event{
		clientID: clientID,
		action:   "has disconnected",
	}

	populateNickname(&clients, &event)

	if event.nickname != nickname {
		t.Fatal("Event is not populated by the nickname")
	}
}

func buildNotification(params NotificationParams) string {
	result := fmt.Sprintf("notifycliententerview cfid=0 ctid=1 reasonid=%s clid=%s client_unique_identifier=%s client_nickname=%s client_input_muted=0 client_output_muted=0 client_outputonly_muted=0 client_input_hardware=1 client_output_hardware=1 client_meta_data client_is_recording=0 client_database_id=2 client_channel_group_id=8 client_servergroups=6 client_away=0 client_away_message client_type=0 client_flag_avatar client_talk_power=75 client_talk_request=0 client_talk_request_msg client_description client_is_talker=0 client_is_priority_speaker=0 client_unread_messages=0 client_nickname_phonetic client_needed_serverquery_view_power=75 client_icon_id=0 client_is_channel_commander=0 client_country client_channel_group_inherited_channel_id=1 client_badges",
		params.reasonid,
		params.clid,
		params.client_unique_identifier,
		params.client_nickname,
	)

	return result
}
