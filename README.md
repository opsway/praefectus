# Boodmo Prӕfectus

### Build and Run
```
go build -o bin/praefectus cmd/*
bin/praefectus run --host="0.0.0.0" --port=9000 --timer-cmd="bin/app cron:process" --timer-interval=60 --worker-number=3 --worker-pool-cmd="bin/app messenger:consume messenger.transport.amqp"
```
