package app

// Module name for logging details
const MODULE = "app"

// Output is marshalled to JSON and sent back to the
// API GW at the end of Lambda function execution.
type Output struct {
	Error    *Err   `json:"error"`
	Status   string `json:"status"`
	ID       string `json:"id"`
	Username string `json:"username"`
}

// Err is a custom error message stuct marshalled
// to JSON and returned by the Lambda function
type Err struct {
	Msg    string `json:"msg"`
	ErrRef string `json:"err_ref"`
}

// AccountCreatedEvent is a struct to reference
// pertinent details sent from Mastodon related
// to the "account.create" event.
type AccoutCreatedEvent struct {
	Event     string      `json:"event"`
	CreatedAt string      `json:"created_at"`
	Object    EventObject `json:"object"`
}

// EventObject contains the Mastodon account details
// required to assess, and if needed, suspend the account.
type EventObject struct {
	Id        string `json:"id"`
	Username  string `json:"username"`
	Domain    string `json:"domain"`
	CreatedAt string `json:"created_at"`
	Email     string `json:"email"`
	Ip        string `json:"ip"`
}

/* Example "account.created" event sent by a Mastondon account.created webhook.
{
	"event": "account.created",
	"created_at": "2023-01-04T01:57:03.708Z",
	"object": {
		"id": "109628451946059725",
		"username": "test001",
		"domain": null,
		"created_at": "2023-01-04T01:57:03.566Z",
		"email": "test001@sigler.io",
		"ip": "73.207.229.36",
		"role": {
			"id": -99,
			"name": "",
			"color": "",
			"position": -1,
			"permissions": 65536,
			"highlighted": false,
			"created_at": "2022-11-18T22:16:44.580Z",
			"updated_at": "2022-11-18T22:16:44.580Z"
		},
		"confirmed": false,
		"suspended": false,
		"silenced": false,
		"sensitized": false,
		"disabled": false,
		"approved": false,
		"locale": "en",
		"invite_request": "I need to test",
		"ips": [{
			"ip": "73.207.229.36",
			"used_at": "2023-01-04T01:57:03.714Z"
		}],
		"account": {
			"id": "109628451946059725",
			"username": "test001",
			"acct": "test001",
			"display_name": "Test001",
			"locked": false,
			"bot": false,
			"discoverable": null,
			"group": false,
			"created_at": "2023-01-04T00:00:00.000Z",
			"note": "",
			"url": "https://nifty-moose.com/@test001",
			"avatar": "https://nifty-moose.com/avatars/original/missing.png",
			"avatar_static": "https://nifty-moose.com/avatars/original/missing.png",
			"header": "https://nifty-moose.com/headers/original/missing.png",
			"header_static": "https://nifty-moose.com/headers/original/missing.png",
			"followers_count": 0,
			"following_count": 0,
			"statuses_count": 0,
			"last_status_at": null,
			"noindex": false,
			"emojis": [],
			"fields": []
		}
	}
}
*/
