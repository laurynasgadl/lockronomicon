# 🔏 Lockronomicon

Lockronomicon is a simple lock service for distributed systems. It provides a slim HTTP API to a FS-based locking mechanism.

## Installation

### From Source
```
> make build
> ./lockronomicon -h
Usage of ./lockronomicon:
  -address string
        Network address to listen on (default ":80")
  -path string
        FS locker workdir path (default "/opt/locker")
  -v    Binary version
```

### Docker
Lockronomicon is available as a [Docker image](https://hub.docker.com/r/laurynasgadl/lockronomicon):
```
docker run -p 80:80 laurynasgadl/lockronomicon
```

## Usage

A lock can be acquired by providing a locking key (pattern [^[\w.-]+$](https://regex101.com/r/IyvYwa/1)) and lock TTL (seconds). successfully acquiring a lock returns its generation number. This number is used to ensure lock ownership.

Lock TTL can be refreshed by providing its key and the generation number - TTL is extended by the original amount.

Locks are released by providing lock key and the generation number.

## API

There are 4 HTTP endpoints in total:

METOD   | URL              | PARAMS     | EXPLANATION
--------|------------------|------------|------------
GET     | /health          |            | A general health check endpoint
POST    | /api/locks       | key, ttl   | For acquiring locks
PUT     | /api/locks/{key} | generation | For refreshing an owned lock
DELETE  | /api/locks/{key} | generation | For releasing an owned lock


### Checking service health
```http
GET /health
```

##### Responses
STATUS | BODY | EXPLANATION
-------|------|------------
200 OK | `{"status":"OK"}` | Service is healthy

##### Example
```bash
> curl localhost:80/health
{"status":"OK"}
```

### Acquiring lock
```http
POST /api/locks
```

##### Params
NAME | TYPE | EXPLANATION
-----|------|------------
key  | string of pattern `^[\w.-]+$` | the lock key
ttl  | int | lock's time-to-live in seconds, negative TTL makes the lock immortal

##### Responses
STATUS | BODY | EXPLANATION
-------|------|------------
200 OK | `{"generation":1622184940255602000}` | Lock acquired successfully
423 Locked | - | Lock already taken

##### Example
```bash
> curl -X POST -H "Content-Type: application/json" -d '{"key":"example.lock_key_1","ttl":300}' localhost:80/api/locks
{"generation":1622283840185146846}
```

### Refreshing lock
```http
PUT /api/locks/{key}
```

##### Params
NAME | TYPE | EXPLANATION
-----|------|------------
key  | string of pattern `^[\w.-]+$` | the lock key
generation  | int | lock's generation number returned upon acquiring it

##### Responses
STATUS | BODY | EXPLANATION
-------|------|------------
200 OK | `{"generation":1622189339302681238}` | Lock refreshed successfully, new generation key returned
412 Precondition Failed | - | Generation number does not match the current one for this lock
404 Not Found | - | Lock with such key does not exist

##### Example
```bash
> curl -X PUT -H "Content-Type: application/json" -d '{"generation":1622283840185146846}' localhost:80/api/locks/example.lock_key_1
{"generation":1622283979363905515}
```

### Releasing lock
```http
DELETE /api/locks/{key}
```

##### Params
NAME | TYPE | EXPLANATION
-----|------|------------
key  | string of pattern `^[\w.-]+$` | the lock key
generation  | int | lock's generation number returned upon acquiring it

##### Responses
STATUS | BODY | EXPLANATION
-------|------|------------
200 OK | - | Lock released successfully
412 Precondition Failed | - | Generation number does not match the current one for this lock
404 Not Found | - | Lock with such key does not exist

##### Example
```bash
> curl -X DELETE -H "Content-Type: application/json" -d '{"generation":1622283979363905515}' localhost:80/api/locks/example.lock_key_1
200 OK
```
