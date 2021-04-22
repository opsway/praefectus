# Boodmo Prӕfectus

### Build and Run
```
go build -o bin/praefectus cmd/*
bin/praefectus run \
    --worker-pool-cmd="bin/app messenger:consume messenger.transport.amqp" \
    --worker-pool-cmd="bin/app messenger:consume messenger.transport.cron" \
    --timer-cmd="bin/app cron:process" --timer-interval=60 \
    --host="0.0.0.0" --port=9000
```
