FROM alpine:latest

RUN mkdir /app

COPY logsApp /app

CMD ["/app/logsApp"]