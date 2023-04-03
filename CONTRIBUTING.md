# Contributing

## Testing
```sh
export FAUNADB_SECRET="faunadb secret"
export CALENDAR_ID="google calendar ID"
export TEST_EMAIL="test email address"
```

The "unit" tests are more of integration tests as they will test with Fauna and Google Calendar.
```sh
go test ./...
```

## Running the server
Create a test config and adjust as needed.
```sh
cp configs/pizza.yaml configs/pizza.test.yaml
```
Start the server
```sh
go run main.go -config configs/pizza.test.yaml
```

## Releasing
1. Test the release
```sh
goreleaser release --snapshot --clean
```
2. Tag the new version
```sh
go tag v0.1.0
```
3. Publish
```sh
goreleaser release --clean
```
