
FROM golang:1.26 AS builder

WORKDIR /build

COPY . .
RUN go build -o orcastack main.go
RUN go build -o orcastack-analytics-service ./cmd/orcastack-analytics-service
RUN go build -o orcastack-device-orch ./cmd/orcastack-device-orch
RUN go build -o orcastack-hw-automation ./cmd/orcastack-hw-automation
RUN go build -o orcastack-secctl ./cmd/orcastack-secctl
RUN go build -o orcastack-cd-service ./cmd/orcastack-cd-service
RUN go build -o orcastack-gateway ./cmd/orcastack-gateway
RUN go build -o orcastack-review-service ./cmd/orcastack-review-service
RUN go build -o orcastack-sw-automation ./cmd/orcastack-sw-automation
RUN go build -o orcastack-ci-service ./cmd/orcastack-ci-service
RUN go build -o orcastack-git-service ./cmd/orcastack-git-service
RUN go build -o orcastack-runner ./cmd/orcastack-runner


FROM debian:stable-slim
WORKDIR /usr/local/bin
COPY --from=builder /build/orcastack .
COPY --from=builder /build/orcastack-analytics-service .
COPY --from=builder /build/orcastack-device-orch .
COPY --from=builder /build/orcastack-hw-automation .
COPY --from=builder /build/orcastack-secctl .
COPY --from=builder /build/orcastack-cd-service .
COPY --from=builder /build/orcastack-gateway .
COPY --from=builder /build/orcastack-review-service .
COPY --from=builder /build/orcastack-sw-automation .
COPY --from=builder /build/orcastack-ci-service .
COPY --from=builder /build/orcastack-git-service .
COPY --from=builder /build/orcastack-runner .


EXPOSE 8080 8081 8082 8083 8084 8085 8086 8087 8088 8089


CMD ["./orcastack", "serve-all"]
