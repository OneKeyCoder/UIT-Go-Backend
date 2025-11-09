FROM gcr.io/distroless/static

COPY api-gateway/gatewayApp /

EXPOSE 80

CMD ["/gatewayApp"]