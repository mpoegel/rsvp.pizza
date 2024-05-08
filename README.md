# rsvp.pizza

rsvp.pizza is a web application for collecting RSVP's to your pizza parties. Your friends can select the days they want to come and enter their email address to receive a Google Calendar invite to the party.

## Installing

Before installing rsvp.pizza, you'll need to get OAuth2 credentials for the Google Calendar API and initialize the database.

### How to get calendar credentials
1. [Start here](https://support.google.com/googleapi/answer/6158849?hl=en&ref_topic=7013279) to create an OAuth Desktop Application.
2. Download the credential and save as `credentials.json`
3. Renew the token `go run cmd/renew_calendar_credentials.go`
4. Copy the printed URL to your web browser and complete the steps to log in with your Google account.
5. Copy the code from the final URL that you're redirected to on localhost that does not exist.

### Initialize the database
```sh
rsvp.pizza patch -init
```

### Setup OAuth2 Server
Configure a Keycloak OAuth2 server. Create a client application to get the Client ID and Client Server.

### Install the package
1. Download the latest version
```sh
wget https://github.com/mpoegel/rsvp.pizza/releases/download/v0.1.0/rsvp.pizza_Linux_x86_64.tar.gz
```
2. Create a pizza user.
```sh
sudo adduser pizza
```
3. Unpack
```sh
sudo tar xzfv rsvp.pizza_Linux_x86_64.tar.gz -C /
```
4. Adjust the environment variables and config file.
```sh
cp /etc/pizza/.env /etc/pizza/.env.prod
sudo vim /etc/pizza/.env.prod
```
5. Adjust the nginx config.
```sh
cp /etc/pizza/nginx.conf /etc/nginx/sites-available/pizza.conf
sudo vim /etc/nginx/sites-available/pizza.conf
sudo ln -s /etc/nginx/sites-available/pizza.conf /etc/nginx/sites-enabled/pizza.conf
sudo systemctl reload nginx
```
6. Start the pizza service.
```sh
sudo systemctl start pizza.service
```
