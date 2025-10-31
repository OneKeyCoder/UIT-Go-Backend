FROM gcr.io/distroless/static

COPY tripServiceApp /

EXPOSE 80
EXPOSE 50054

CMD ["/tripServiceApp"]