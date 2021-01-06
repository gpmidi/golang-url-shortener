FROM golang:1.15-alpine AS BUILD

# Deps
RUN apk update && apk add --no-cache ca-certificates curl yarn bash tcpdump vim git
RUN go get -v github.com/mattn/goveralls \
 && go get -v github.com/gobuffalo/packr/v2/packr2 \
 && go get -v github.com/mitchellh/gox

COPY . /go/src/github.com/mxschmitt/golang-url-shortener/
WORKDIR /go/src/github.com/mxschmitt/golang-url-shortener/
RUN go mod tidy && go mod download && go mod tidy && go mod vendor


# FE
RUN cd web && yarn install && yarn build && rm build/static/**/*.map

# Build it
RUN go build cmd/golang-url-shortener/main.go



FROM alpine AS RUN

LABEL maintainer="Paulson McIntyre <paul@gpmidi.net>"
LABEL readme.md="https://github.com/mxschmitt/golang-url-shortener/blob/master/README.md"
LABEL description="This Dockerfile will install the Golang URL Shortener."

RUN apk update && apk add --no-cache ca-certificates curl bash tcpdump vim git

COPY --from=BUILD /go/src/github.com/mxschmitt/golang-url-shortener/main /usr/bin/golang-url-shortener
COPY --from=BUILD /go/src/github.com/mxschmitt/golang-url-shortener/web/build/static /static
COPY --from=BUILD /go/src/github.com/mxschmitt/golang-url-shortener/internal/handlers/tmpls /tmpls
COPY --from=BUILD /go/src/github.com/mxschmitt/golang-url-shortener/web /web

HEALTHCHECK --interval=30s CMD curl -f http://127.0.0.1:$PORT/api/v1/info || exit 1

CMD ["/usr/bin/golang-url-shortener"]
