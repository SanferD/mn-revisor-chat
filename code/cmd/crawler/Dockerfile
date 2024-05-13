# first stage: certs
FROM alpine:latest as certs
RUN apk --update add ca-certificates

# final stage: image with certs and deployable binary
FROM scratch
ARG BINARY_PATH
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ${BINARY_PATH} /
CMD ["/crawler"]