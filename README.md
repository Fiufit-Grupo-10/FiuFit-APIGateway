![ci](https://github.com/Fiufit-Grupo-10/FiuFit-APIGateway/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/Fiufit-Grupo-10/FiuFit-APIGateway/branch/main/graph/badge.svg?token=CQMMLS2MR5)](https://codecov.io/gh/Fiufit-Grupo-10/go-template)
# go-template
Template for go services

### Ejecutar servidor en local

```bash
# Parado sobre la carpeta principal del proyecto
docker-compose up
```

### Verificar que se haya levantado

```bash
# En una nueva terminal
curl localhost:8080/ping
```

```bash
# Respuesta esperada
{"message":"pong"}
```

### Para trabajar dentro del container corriendo

```bash
# Ejecutar docker ps para obtener el ID del container y ejecutar

docker exec -it <ID> bash

# se deberia levantar un proceso con una terminal dentro del container, ya se pueden ejecutar tests.
```
