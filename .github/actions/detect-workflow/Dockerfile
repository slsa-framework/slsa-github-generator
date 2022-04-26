FROM golang@sha256:3b7fa1bed1510cfdadbbd839ab1c2c4ca5ee86bc259d2cb71142f6bd4e70b59f as builder

WORKDIR /app
COPY . /app

RUN go get -d -v

# Statically compile our app for use in a distroless container
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app .

# A distroless container image with some basics like SSL certificates
# https://github.com/GoogleContainerTools/distroless
FROM gcr.io/distroless/static@sha256:0bec8e882c023de03ba9ed6bb07cd6b5e52b19fb1641505ea1b1c8937f3ccc7d

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]