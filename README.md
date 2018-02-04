# GOChat

## API

### User

```
GET /user
```

Get a list of user

```
GET /user/{userID}
```

Get the info of user

### Register

```
POST /register


Multipart Form:
- username string
- password string
```

Create a new user

### Auth

```
GET /auth

Multipart Form:
- userid int
- password string
```

Returns a JWT token

## Protobufs

Generate Protobuf files using:

```
protoc --go_out=plugins=grpc:. *.proto
```
