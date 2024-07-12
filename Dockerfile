FROM golang

WORKDIR /usr/src/pizza

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY static /etc/pizza/static

COPY . .
RUN go build -o /usr/local/bin/rsvp.pizza

ENV PIZZA_STATIC_DIR=/etc/pizza/static
ENV CREDENTIAL_FILE=/etc/pizza/creds/credentials.json
ENV TOKEN_FILE=/etc/pizza/creds/token.json
ENV DBFILE=/etc/pizza/db/pizza.db
ENV PORT=9090
ENV METRICS_PORT=9191

EXPOSE 9090
EXPOSE 9191

VOLUME [ "/etc/pizza/db/", "/etc/pizza/creds/" ]

CMD ["/usr/local/bin/rsvp.pizza", "run"]
