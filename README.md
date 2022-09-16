### HOW TO RUN
To run the application, you'll need to make sure you have Docker and Docker compose installed

You can start the application by running the following command:

```bash
make start
```

This will build the docker image and start the container in detached mode.

To stop the container, run the following command: 

```bash
make stop
```

### API
- Create Answer

```bash
curl --location --request POST 'http://localhost:5005/api/v1/answers' \
--header 'Content-Type: application/json' \
--data-raw '{
    "key": "1234567",
    "value": "new-answer"
}'
```

- Get Answer

```bash
curl --location --request GET 'http://localhost:5005/api/v1/answers/1234567'
```

- Update Answer

```bash
curl --location --request PUT 'http://localhost:5005/api/v1/answers/1234567' \
--header 'Content-Type: application/json' \
--data-raw '{
    "value": "tech-nation-key-update"
}'
```

- Delete Answer

```bash
curl --location --request DELETE 'http://localhost:5005/api/v1/answers/123456'
```

- Get History by Key

```bash
curl --location --request GET 'http://localhost:5005/api/v1/answers/1234567/history?perPage=20&page=1'
```



### Testing 
To run integration tests, you'll need to make sure `TEST_MONGO_DSN` is set as an environment variable and points to your Test DB instance. You can run integration tests by running the following command:

```go
go test -tags integration -p 1 ./...
```

To run unit tests, you can run the following command:

```go
go test -v ./...
```


## Questions

1. How would you support Multiple users?
  - We can support multiple users by scoping an answer a to user by introducing a new `user_id` field to the `answers` collection. This means there would be also be the need
  to provide support for  authentication and authorization which can be in the form of `jwt` tokens and api Keys.

2. How would you support answers with types other than string?
 - 

3. What are the main bottlenecks of your solution?
 - The main bottleneck of my solution will be the Database. Currently, every request requires a read/write query to the Database which can further increase latency when there are thousands of concurrent requests happening at the same time.

4. How would you scale the service to cope with thousands of requests?
- By introducing a caching strategy using Redis. With a cache in place, not every read request needs to be routed to the Database, some of the keys can be fetched from the cache which will help to improve latency and throughput.
- By Horizontal scaling, having multiple instances of the service running behind a LoadBalancer.