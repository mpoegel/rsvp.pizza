# rsvp.pizza

## How to get calendar credentials
1. [Start here](https://support.google.com/googleapi/answer/6158849?hl=en&ref_topic=7013279) to create an OAuth Application.
2. Download the credential and save as `credentials.json`
3. Renew the token `go run cmd/renew_calendar_credentials.go`

## Required Environment Variables

### Runtime
```sh
export FAUNADB_SECRET="faunadb secret"
```

### Testing
```sh
export CALENDAR_ID="google calendar ID"
export TEST_EMAIL="test email address"
```
