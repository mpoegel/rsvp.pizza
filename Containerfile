
FROM docker.io/debian:trixie-slim
RUN apt-get update && apt-get install -y --no-install-recommends libsqlite3-0 ca-certificates libgpgme11t64 && \
    rm -rf /var/lib/apt/lists/*
COPY static /etc/rsvp.pizza/static
ENV PIZZA_STATIC_DIR=/etc/rsvp.pizza/static
ARG BINARY=rsvp.pizza
COPY ${BINARY} /rsvp.pizza
ENTRYPOINT [ "/rsvp.pizza" ]
