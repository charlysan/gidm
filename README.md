# Guy In The Middle (gidm)
Simple man-in-the-middle tool

## Intro

`gidm` is a minimalist "man-in-the-middle" tool that can be used as a forward proxy capable of injecting custom headers and logging requests

## Requirements

`GO v1.15` or greater 

## Installation

Just clone this repo and build & install as usual:

```
go install
```

## Usage

You can run the tool with `--help` option to get a list of supported commands:

```bash
$ gidm --help
NAME:
   gidm - Simple midm tool

USAGE:
   gidm [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -H value    inject header to request
   -P value    listen on port (default: "8080")
   -U value    redirect to url (default: "http://localhost:9000")
   -d          enable debugging (default: false)
   --help, -h  show help (default: false)
```

## Examples

Listen on port `8081` and forward to `http://localhost:9000`; show debug information and inject the following headers:
- `x-custom-flag: true`
- `x-custom-id: 12345`


```bash
gidm \
-P 8081 \
-U http://localhost:9000 \
-H "x-custom-flag: true" \
-H "x-custom-id: 12345" \
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
Headers to be injected:
  x-custom-flag: true
  x-custom-id: 12345

2021/05/25 21:31:20 POST HTTP/1.1 localhost:8081/dummy

{"name": "john doe"}

Content-Length: [20]
User-Agent: [curl/7.64.1]
Accept: [*/*]
Content-Type: [application/json]
```

If you want to check that the headers are being injected, you could run two `gidm` instances and chain them together:

```
gidm \
-P 8081 \
-U http://localhost:8082 \
-H "x-custom-flag: true" \
-H "x-custom-id: 12345" \
-d
```

```
gidm \
-P 8082 \
-U http://localhost:9000 \
-d
```

And in the second instance you should see something like this:
```
Listening on port: 8082
Redirecting to: http://localhost:9000

2021/05/25 21:38:29 POST HTTP/1.1 localhost:8082/dummy

{"name": "john doe"}

Content-Length: [20]
Accept: [*/*]
Content-Type: [application/json]
X-Forwarded-For: [::1]
Accept-Encoding: [gzip]
User-Agent: [curl/7.64.1]
X-Custom-Id: [12345]
X-Custom-Flag: [true]
```
