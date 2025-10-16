#### Leaked passwords API

##### Description
API for checking if a password is leaked in breaches. Uses [HIBP](https://haveibeenpwned.com/API/v3) API.

#### ADR
[ADR-001](adr/adr-001-storage-strategy-for-pwned-passwords.md): Storage Strategy for Pwned Password Dataset

#### Development
```shell
APP_ENV=development go run src/main.go
```

#### Running in a Docker container
```shell
docker build -t pwned-api .
docker run --env-file .env.development -e APP_ENV=development -e RUNNING_IN_DOCKER=true -p 8080:8080 pwned-api
```

#### Want to see it in action with the Spring authorization server?
Go to this [Github repo](https://github.com/GoodbyePlanet/spring-cg-bff) and follow readme.

Example CURL
```shell
curl -X POST http://localhost:8080/check \
  -H "Content-Type: application/json" \
  -d '{"password": "movies15"}'
```
