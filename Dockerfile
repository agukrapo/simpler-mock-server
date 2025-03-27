FROM golang:1 as builder
WORKDIR /go/src/github.com/agukrapo/simpler-mock-server
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sms ./cmd/sms

FROM alpine:3
WORKDIR /app
COPY --from=builder /go/src/github.com/agukrapo/simpler-mock-server/sms .
COPY --from=builder /go/src/github.com/agukrapo/simpler-mock-server/.sms_responses ./.sms_responses
ENTRYPOINT ["./sms"]
