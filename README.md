![ci](https://github.com/Fiufit-Grupo-10/FiuFit-APIGateway/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/Fiufit-Grupo-10/FiuFit-APIGateway/branch/main/graph/badge.svg?token=CQMMLS2MR5)](https://codecov.io/gh/Fiufit-Grupo-10/FiuFit-APIGateway)
# Fiufit API Gateway
API Gateway implementation for the Fiufit application. It acts as an
intermediary between the frontend applications (Backoffice & Mobile)
and the backend.

### Running dev enviroment
The next command runs the application with live-reloading usign [cosmtrek/air](https://github.com/cosmtrek/air). 
This makes the development process more intereactive

```bash
$ make air
```
To run tests,
```bash
$ make tests
```
Or if running inside a container is prefered
```bash
$ make docker-test
```

### Building
The next command builds a native binary named main
```bash
$ make build
```
Container version:
```bash
$ make build-docker
```

### License
[LICENSE-MIT](https://opensource.org/license/mit/)
