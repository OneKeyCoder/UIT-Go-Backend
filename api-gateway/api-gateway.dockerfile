FROM gcr.io/distroless/static

COPY gatewayApp /

EXPOSE 80

CMD ["/gatewayApp"]