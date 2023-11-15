# be-service

a simple backend service that returns random kanye west quotes with simulated latency

the backend has one endpoint:
```
GET /kanye
```

you can add latency by passing the delay parameter, this will add a random delay of up to 3 seconds:

```
GET /kanye?delay=true
```

## usage
### running locally
```sh
go run quoter.go
```

### running on docker 
build the image
```sh
docker build -t quoter .
``````

now you can run the container like so
```sh
docker run -p 8080:8001 quoter
```

or use the docker-compose to bring up multiple backends at once
```sh
docker compose up -d
```

## getting the quotes
you can now access the backend at http://localhost:8001/kanye
```sh
curl 'localhost:8001/kanye?delay=false'
```
