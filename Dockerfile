FROM scratch
COPY rsvp.pizza /
ENTRYPOINT [ "/rsvp.pizza" ]
COPY static /etc/rsvp.pizza/static
ENV STATIC_DIR=/etc/rsvp.pizza/static
