FROM gcr.io/distroless/static

COPY authApp /

EXPOSE 80
EXPOSE 50051

CMD ["/authApp"]