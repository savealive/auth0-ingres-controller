FROM alpine:3.9
RUN apk add --update ca-certificates

ENTRYPOINT ["/auth0-ingress-controller"]
COPY ./bin/auth0-ingress-controller /