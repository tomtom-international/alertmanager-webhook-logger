FROM golang:1.18 as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -installsuffix 'static' . 

RUN echo 'nobody:x:65534:65534:nobody:/:' > /etc/passwd && \
    echo 'nobody:x:65534:' > /etc/group


FROM scratch

COPY --from=builder /app/alertmanager-webhook-logger /app/alertmanager-webhook-logger
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Perform any further action as an unprivileged user.
USER nobody:nobody

ENTRYPOINT ["/app/alertmanager-webhook-logger"]
