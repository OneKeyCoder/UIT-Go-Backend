FROM gcr.io/distroless/static

COPY authentication-service/authApp /

EXPOSE 80
EXPOSE 50051

CMD ["/authApp"]