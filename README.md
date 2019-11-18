# Alertmanager Webhook Logger

Generates (structured) log messages from Prometheus AlertManager WebHooks.

## Rationale

The [Prometheus Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) does not provide any history of alerts. Logging of alerts is the most simple solution to get that history. In combination with log management solutions like [Elastic Stack](https://www.elastic.co/products/), etc. it should fit most use-cases for a comfortable history of alerts.

## Usage

Command line flags:
    .\alertmanager-webhook-logger -h

# TODO repo
    go get -u github.com/TODO
    cd $env:GOPATH/src/github.com/TODO
    .\alertmanager-webhook-logger

## License

Under [MIT](LICENSE)