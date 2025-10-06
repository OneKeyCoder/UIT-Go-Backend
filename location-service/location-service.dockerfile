FROM alpine:latest

RUN mkdir /app

COPY locationServiceApp /app

CMD ["/app/locationServiceApp"]
