# GOChat

## API

Generate API Documentation using `apidoc`

## Protobufs

Generate Protobuf files using:

```
protoc --go_out=plugins=grpc:. *.proto
```

## Ports

* `:3000`: website
* `:5432`: postgresql
* `:6379`: redis
* `:8000`: traefik front-end
* `:8080`: traefik dashboard
* `:9000`: portainer
* `:9042`: cassandra1
