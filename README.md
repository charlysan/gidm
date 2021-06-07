# gidm
Simple http proxy man-in-the-middle tool

## Intro

`gidm` is a minimalist http proxy that can be used as "man-in-the-middle" tool capable of manipulating requests and responses.

## Installation

### Build from source

[GO v1.15](https://golang.org/doc/go1.15) or greater is required. Just clone this repo and build as usual:

```
go build
```

Or build & install:

```
go install
```

### Download release binaries

You can also download Linux and MacOS pre-built binaries from [releases section](https://github.com/charlysan/gidm/releases).

### Docker Image

Docker image is also available from [DockerHub](https://hub.docker.com/repository/docker/charlysan/gidm). You will find some examples in [Examples Section](#run-using-docker).


## Usage

You can run the tool along with `--help` option to get a list of supported commands:

```bash
$ ./gidm --help
NAME:
   gidm - Simple http proxy midm tool

USAGE:
   gidm [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --reqh value  inject request header
   --resh value  inject response header
   --reqb value  replace string in request body (/old/new/)
   --resb value  replace string in response body (/old/new/)
   -p value      listen to port (default: "8080")
   -u value      redirect to url (default: "http://localhost:9000")
   -i value      enable interactive mode (API server will listen on specified port)
   -d            enable debugging (default: false)
   --help, -h    show help (default: false)
```

## Examples

### Inject Request Headers

Listen on port `8081` and forward to `http://localhost:9000`, show debug information and inject the following headers:
- `x-custom-flag: true`
- `x-custom-id: 12345`


```bash
gidm \
-p 8081 \
-u http://localhost:9000 \
-reqh "x-custom-flag: true" \
-reqh "x-custom-id: 12345" \
-d
```

POST something to `localhost:8081`:
```
curl -X POST \
http://localhost:8081/dummy \
-H "content-type: application/json" \
-d '{"name": "john doe"}'
```

You should get this output:
```
Listening on port: 8081
Redirecting to: http://localhost:9000

Request headers to be injected:
  x-custom-flag: true
  x-custom-id: 12345
2021/05/31 18:34:16 POST /dummy HTTP/1.1
Host: localhost:9000
Accept: */*
Content-Length: 20
Content-Type: application/json
User-Agent: curl/7.64.1
X-Custom-Flag: true
X-Custom-Id: 12345

{"name": "john doe"}
2021/05/31 18:34:16 HTTP/1.1 404 Not Found
Content-Length: 22
Content-Type: application/json
Date: Mon, 31 May 2021 21:34:16 GMT
Server: uvicorn

{"detail":"Not Found"}
```

### Replace body strings

You can add string replacers for request and response body. 
For example, to replace every `ok` with `BAD` in your response body, you can use this command:

```bash
./gidm \
-p 8081 \
-u http://localhost:9000 \
-reqh "x-custom-flag: true" \
-reqh "x-custom-id: 12345" \
-resb "/ok/BAD/" \
-d
```

## Interactive Mode

Interactive Mode allows to modify the proxy behavior without restarting the app. Te proxy will listen on the port specified using `-i <PORT>` flag:

```bash
./gidm \
-p 8081 \
-u http://localhost:9000 \
-reqh "x-custom-flag: true" \
-reqh "x-custom-id: 12345" \
-resb "/ok/BAD/" \
-i 9090 \
-d
```

And it will expose the following endpoints:

```
PUT /requestHeaders
PUT /responseHeaders
PUT /requestBodyReplacers
PUT /responseBodyReplacers
```

So, supposing you want to change the response body string replacers you can hit the proxy with this payload:

```
curl -X PUT \
http://localhost:9090/responseBodyReplacers \
-d '{"ok": "WRONG!!"}'
```

And the proxy should show the following log:
```
Response Body string replacers updated
  ok -> WRONG!!
```

## Run using Docker

Grab Docker image from [DockerHub](https://hub.docker.com/repository/docker/charlysan/gidm):

```
docker pull charlysan/gidm
```

Run and add the proper port-forwarding:

```
docker run \
-p 8081:8080 \
-p 9090:9090 \
charlysan/gidm \
-u https://api.chucknorris.io \
-resb "/Chuck Norris/John Doe/" \
-i 9090 \
-d
```
