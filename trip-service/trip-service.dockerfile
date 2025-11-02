FROM gcr.io/distroless/static

COPY trip-service/tripServiceApp /

EXPOSE 80
EXPOSE 50054

CMD ["/tripServiceApp"]