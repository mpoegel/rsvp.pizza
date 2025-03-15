#!/bin/bash

set -e

ADDR=${ADDR:-"http://localhost:9090"}

echo ">> Authorizing"
JWT=$(curl -s -X POST $ADDR/api/token \
    -H 'Content-Type: application/x-www-form-urlencoded' \
    -d username=$USERNAME \
    -d password=$PASSWORD \
    -d grant_type=password)

export TOKEN=$(echo $JWT | jq -r .access_token)

echo ">> Fetching fridays"
fridays=$(curl -s -X GET $ADDR/api/friday \
    -H 'Accept: application/vnd.api+json' \
    -H "Authorization: Bearer $TOKEN")
jq . <<< "$fridays"

friday_id=$(jq -r .data[0].id <<< $fridays)

echo ">> RSVPing to $friday_id"
new_friday=$(curl -s -X PATCH $ADDR/api/friday/$friday_id \
    -H 'Content-Type: application/vnd.api+json' \
    -H 'Accept: application/vnd.api+json' \
    -H "Authorization: Bearer $TOKEN" \
    --data "{\"data\": { \"type\": \"friday\", \"id\": \"$friday_id\", \"relationships\": { \"guests\": { \"data\": [{ \"type\": \"guest\", \"id\": \"1\" }] } } } }")

echo ">> RSVP complete"
jq . <<< "$new_friday"

echo ">> Fetching guest"
guest=$(curl -s -X GET $ADDR/api/guest/1 \
    -H 'Accept: application/vnd.api+json' \
    -H "Authorization: Bearer $TOKEN")
jq . <<< "$guest"

echo ">> Fetching guest profile"
profile=$(curl -s -X GET $ADDR/api/guest/1/profile \
    -H 'Accept: application/vnd.api+json' \
    -H "Authorization: Bearer $TOKEN")
jq . <<< "$profile"
