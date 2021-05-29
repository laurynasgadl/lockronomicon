# Lockronomicon

Lockronomicon is a simple centralized lock system for distributed services. It provides a slim HTTP API to a FS-based locking mechanism.

## Usage

A lock can be acquired by providing a locking key (pattern [^[\w.-]+$](https://regex101.com/r/IyvYwa/1)) and lock TTL (seconds). Succesfully acquiring a lock returns its generation number. This number is used to ensure lock ownership.

Lock TTL can be refreshed by providing its key and the generation number - TTL is extended by the original amount.

Locks are released by providing lock key and the generation number.

### API

There are 4 HTTP endpoints in total:

METOD   | URL              | PARAMS     | EXPLANATION
--------|------------------|------------|------------
GET     | /health          |            | A general health check endpoint
POST    | /api/locks       | key, ttl   | For acquiring locks
PUT     | /api/locks/{key} | generation | For refreshing an owned lock
DELETE  | /api/locks/{key} | generation | For releasing an owned lock


#### Checking service health
##### Request
```http
GET /health
```
##### Response
```json
{
    "status":"OK"
}
```

#### Acquiring lock
##### Request
```http
POST /api/locks
{
    "key":"example.lock_key_1",
    "ttl":300,
}
```
##### Params
NAME | TYPE | EXPLANATION
-----|------|------------
key  | string of pattern `^[\w.-]+$` | the lock key
ttl  | int | lock's time-to-live in seconds

##### Responses
Succesfully acquiring a lock returns its generation key:
```json
{
    "generation":1622184940255602000
}
```
If a lock with such key is already taken, HTTP error with status code 423 will be returned:
```http
423 Locked
```

#### Refreshing lock
##### Request
```http
PUT /api/locks/{key}
{
    "generation":1622184940255602000
}
```
##### Params
NAME | TYPE | EXPLANATION
-----|------|------------
key  | string of pattern `^[\w.-]+$` | the lock key
generation  | int | lock's generation number returned upon acquiring it

##### Responses
Succesfully refreshing a lock returns a new generation key:
```json
{
    "generation":1622189339302681238
}
```
If the provided generation number for the lock does not match the current one, HTTP 412 will be returned:
```http
412 Precondition Failed
```
If such lock does not exist, HTTP 404 will be returned:
```http
404 Not Found
```

#### Releasing lock
##### Request
```http
DELETE /api/locks/{key}
{
    "generation":1622184940255602000
}
```
##### Params
NAME | TYPE | EXPLANATION
-----|------|------------
key  | string of pattern `^[\w.-]+$` | the lock key
generation  | int | lock's generation number returned upon acquiring it

##### Responses
Succesfully releasing a lock returns HTTP 200:
```http
200 OK
```
If the provided generation number for the lock does not match the current one, HTTP 412 will be returned:
```http
412 Precondition Failed
```
If such lock does not exist, HTTP 404 will be returned:
```http
404 Not Found
```