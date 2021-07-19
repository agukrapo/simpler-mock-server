FROM golang:1 as builder
WORKDIR /go/src/github.com/agukrapo/simpler-mock-server
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sms ./cmd

FROM alpine
COPY --from=builder /go/src/github.com/agukrapo/simpler-mock-server/sms .
COPY --from=builder /go/src/github.com/agukrapo/simpler-mock-server/responses ./responses
COPY --from=builder /go/src/github.com/agukrapo/simpler-mock-server/content-type-mapping.txt .
ENTRYPOINT ["./sms"]