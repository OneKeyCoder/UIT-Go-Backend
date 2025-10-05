FROM gcr.io/distroless/static

COPY loggerServiceApp /

EXPOSE 80
EXPOSE 50052

CMD ["/loggerServiceApp"]