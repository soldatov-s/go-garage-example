# Example project with go-garage

The example of project, which created on base go-garage. Everything is unified as much as possible, no global variables.
The example supported work with rabbitmq, with redis, database (postgresql) with migrations, prometheus, swagger 2.0.
For unification, all work with the project is placed in the makefile (see help for commands, just run make for this).

## How starts
```bash
make docker-compose-up
```
Will be started:
* postgresql
* redis
* rabbitmq
* go-garage-example service  

Service applies postgresql migrations by self.  
After starting you can find swaggers:
* http://localhost:9000/api/v1/swagger/index.html
* http://localhost:9100/api/v1/swagger/index.html  

Prometheus metrics http://localhost:9100/metrics  
Alive http://localhost:9100/health/alive  
Ready http://localhost:9100/health/ready  

You can test sending messages to rabbitmq and consuming messages from it and caching data in redis.
For testing rabbitmq:
- create record

```go
curl -X 'POST' \
  'http://localhost:9100/api/v1/test' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "code": "test",
  "id": 1
}'
```

- post message {"code":"test","send_at":1641372157} with route key TEST_EVENTS to Exchange: test.events.dev via rabbitMQ WebUI
- watch message at Exchange: testout.events.dev