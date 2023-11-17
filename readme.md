# load-balancer

<img src='demo-server.gif' width='300'>
<img src='demo-client.gif' width='300'>


a load balancer service written in Go
* provides load balancing across multiple backend services using 
  - round robin 
  - least connected algorithm 
* configurable through a config file

## usage
### configuration
The load balancer is configured via a config.yaml file:

```yaml
lb_port: 8005
strategy: round-robin
backends:
  - "http://localhost:8001"
  - "http://localhost:8002"
  - "http://localhost:8003"
  - "http://localhost:8004"
health_check_interval_in_sec: 30
max_retries: 5
```

### running
```bash
go run main.go
```

### sample backend service
a sample backend service has been provided in [be-service](be-service) with its own docker-compose to bring up multiple servies