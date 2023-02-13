# Caddy-NATS-Bridge

> NOTE: This package is originally based on https://github.com/codegangsta/caddy-nats,
> which has greatly inspired me to work on this topic. Because we want to use this
> package in our core infrastructure very heavily and I had some specific ideas
> around the API, I created my own package based on the original one.
> 
> tl;dr: Pick whichever works for you - Open Source rocks :)

`caddy-nats-bridge` is a caddy module that allows the caddy server to interact with a
[NATS](https://nats.io/) server. This extension supports multiple patterns:
publish/subscribe, fan in/out, and request reply.

The purpose of this project is to better bridge HTTP based services with NATS
in a pragmatic and straightforward way. If you've been wanting to use NATS, but
have some use cases that still need to use HTTP, this may be a really good
option for you.

<!-- TOC -->
* [Caddy-NATS-Bridge](#caddy-nats-bridge)
  * [Installation](#installation)
  * [Getting Started](#getting-started)
  * [Connecting to NATS](#connecting-to-nats)
  * [Connectivity Modes](#connectivity-modes)
  * [Subscribing to a NATS subject](#subscribing-to-a-nats-subject)
    * [Subscribe Placeholders](#subscribe-placeholders)
    * [subscribe](#subscribe)
      * [Syntax](#syntax)
      * [Example](#example)
    * [reply](#reply)
      * [Syntax](#syntax-1)
      * [Example](#example-1)
    * [queue_subscribe](#queuesubscribe)
      * [Syntax](#syntax-2)
      * [Example](#example-2)
    * [queue_reply](#queuereply)
      * [Syntax](#syntax-3)
      * [Example](#example-3)
  * [Publishing to a NATS subject](#publishing-to-a-nats-subject)
    * [Publish Placeholders](#publish-placeholders)
    * [nats_publish](#natspublish)
      * [Syntax](#syntax-4)
      * [Example](#example-4)
      * [large HTTP payloads with store_body_to_jetstream](#large-http-payloads-with-storebodytojetstream)
    * [nats_request](#natsrequest)
      * [Syntax](#syntax-5)
      * [Example](#example-5)
      * [Format of the NATS message](#format-of-the-nats-message)
  * [Concept](#concept)
  * [What's Next?](#whats-next)
<!-- TOC -->

## Concept Overview

![](./connectivity-modes.drawio.png)

The module works if you want to bridge *HTTP -> NATS*, and also *NATS -> HTTP* - both in unidirectional, and in
bidirectional mode. 

## Installation

To use `caddy-nats-bridge`, simply run the [xcaddy](https://github.com/caddyserver/xcaddy) build tool to create a
`caddy-nats-bridge` compatible caddy server.

```sh
# Prerequisites - install go and xcaddy
brew install go
go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest

# now, build your custom Caddy server (NOTE: path to xcaddy might differ on your system).
~/go/bin/xcaddy build --with github.com/sandstorm/caddy-nats-bridge
```

## Getting Started

Getting up and running with `caddy-nats-bridge` is pretty simple:

First [install NATS](https://docs.nats.io/running-a-nats-service/introduction/installation) and make sure the NATS server is running:

```sh
nats-server
```


> :bulb: To try this example, `cd examples/nats-to-http-request; ./build-run.sh` 

Then create your Caddyfile:

```nginx
{
  nats {
    url 127.0.0.1:4222
    clientName "My Caddy Server"
    
    # if something is received on 
    subscribe hello GET https://localhost/datausa
  }
}

localhost {
  route /datausa {
    https://datausa.io/api/data?drilldowns=Nation&measures=Population
    respond "Hello, world"
  }
}
```

After running your custom built `caddy` server with this Caddyfile, you can
then run the `nats` cli tool to test it out:

```sh
nats req hello ""
```

## Connecting to NATS

To connect to `nats`, simply use the `nats` global option in your Caddyfile with the URL of the NATS server:

```nginx
{
  nats [alias] {
    url nats://127.0.0.1:4222
  } 
}
```

The `alias` is a server-reference which is relevant if you want to connect to two NATS servers at the same time.

To connect to multiple servers (NATS Cluster), separate them with `,` in the `url` parameter.

On top, the following options are supported:

```nginx
{
  nats [alias] {
    url nats://127.0.0.1:4222
    # either userCredentialFile or nkeyCredentialFile can be specified. If both are specified, userCredentialFile
    # takes precedence.
    userCredentialFile /path/to/file.creds
    nkeyCredentialFile /path/to/file.nk
    clientName MyClient
    inboxPrefix _INBOX_custom
  }
}
```

TODO EXPLAIN

## Subscribing to a NATS subject

`caddy-nats` supports subscribing to NATS subjects in a few different flavors,
depending on your needs.

### Subscribe Placeholders

All `subscribe` based directives (`subscribe`, `reply`, `queue_subscribe`, `queue_reply`) support the following `caddy` placeholders in the `method` and `url` arguments:

- `{nats.path}`: The subject of this message, with dots "." replaced with a slash "/" to make it easy to map to a URL.
- `{nats.path.*}`: You can also select a segment of a path ex: `{nats.path.0}` or a subslice of a path: `{nats.path.2:}` or `{nats.path.2:7}` to have ultimate flexibility in what to forward onto the URL path.

---

### subscribe

#### Syntax
```nginx
subscribe <subject> <method> <url>
```

`subscribe` will subscribe to the specific NATS subject (wildcards are
supported) and forward the NATS payload to the specified URL inside the caddy
web server. This directive does not care about the HTTP response, and is
generally fire and forget.

#### Example

Subscribe to an event stream in NATS and call an HTTP endpoint:

```nginx
{
  nats {
      subscribe events.> POST https://localhost/nats_events/{nats.path.1:}
  }
}
```

---

### reply

#### Syntax
```nginx
reply <subject> <method> <url>
```

`reply` will subscribe to the specific NATS subject and forward the NATS
payload to the specified URL inside the caddy web server. This directive will
then respond back to the nats message with the response body of the HTTP
request.

#### Example

Respond to the `hello.world` NATS subject with the response of the `/hello/world` endpoint.

```nginx
{
  nats {
    reply hello.world GET https://localhost/hello/world
  }
}
```

---

### queue_subscribe

#### Syntax
```nginx
queue_subscribe <subject> <queue> <method> <url>
```
`queue_subscribe` operates the same way as `subscribe`, but subscribes under a NATS [queue group](https://docs.nats.io/nats-concepts/core-nats/queue)

#### Example

Subscribe to a worker queue:

```nginx
{
  nats {
    queue_subscribe jobs.* workers_queue POST https://localhost/{nats.path}
  }
}
```

---

### queue_reply

#### Syntax
```nginx
queue_reply <subject> <queue> <method> <url>
```

`queue_reply` operates the same way as `reply`, but subscribes under a NATS [queue group](https://docs.nats.io/nats-concepts/core-nats/queue)

#### Example

Subscribe to a worker queue, and respond to the NATS message:

```nginx
{
  nats {
    queue_reply jobs.* workers_queue POST https://localhost/{nats.path}
  }
}
```

## Publishing to a NATS subject

`caddy-nats` also supports publishing to NATS subjects when an HTTP call is
matched within `caddy`, this makes for some very powerful bidirectional
patterns.

### Publish Placeholders

All `publish` based directives (`nats_publish`, `nats_request`) support the following `caddy` placeholders in the `subject` argument:

- `{nats.subject}`: The path of the http request, with slashes "/" replaced with dots "." to make it easy to map to a NATS subject.
- `{nats.subject.*}`: You can also select a segment of a subject ex: `{nats.subject.0}` or a subslice of the subject: `{nats.subject.2:}` or `{nats.subject.2:7}` to have ultimate flexibility in what to forward onto the URL path.

Additionally, since `publish` based directives are caddy http handlers, you also get access to all [caddy http placeholders](https://caddyserver.com/docs/modules/http#docs).

---

### nats_publish

#### Syntax
```nginx
nats_publish [<matcher>] <subject> {
  timeout <timeout-ms>
}
```
`nats_publish` publishes the request body to the specified NATS subject. This
http handler is not a terminal handler, which means it can be used as
middleware (Think logging and events for specific http requests).

#### Example

Publish an event before responding to the http request:

```nginx
localhost {
  route /hello {
    nats_publish events.hello
    respond "Hello, world"
  }
}
```

#### large HTTP payloads with store_body_to_jetstream

(TODO describe here)

```nginx
localhost {
  route /hello {
    store_body_to_jetstream // TODO: condition if message is large or chunked.
    nats_publish events.hello
    respond "Hello, world"
  }
}
```

---

### nats_request

#### Syntax
```nginx
nats_request [<matcher>] <subject> {
  timeout <timeout-ms>
}
```
`nats_request` publishes the request body to the specified NATS subject, and
writes the response of the NATS reply to the http response body.

#### Example

Publish an event before responding to the http request:

```nginx
localhost {
  route /hello/* {
    nats_request hello_service.{nats.subject.1}
  }
}
```

#### Format of the NATS message

- HTTP Body = NATS Message Data
- HTTP Headers = NATS Message Headers
  - `X-NatsBridge-Method` header: contains the HTTP header `GET,POST,HEAD,...`
  - `X-NatsBridge-UrlPath` header: URI path without query string
  - `X-NatsBridge-UrlQuery` header: query string
  - `X-NatsBridge-LargeBody-Bucket` header
  - `X-NatsBridge-LargeBody-Id` header
- NATS messages have a size limit of usually 1 MB (and 8 MB as hardcoded limit).
  In case the HTTP body is bigger, or alternatively, is submitted with `Transfer-Encoding: chunked` (so we do not know the size upfront);
  we do the following:
  - We store the HTTP body in the [JetStream Object Storage (EXPERIMENTAL)](https://docs.nats.io/using-nats/developer/develop_jetstream/object)
    for a few minutes; in a random key.
  - The name of this KV Storage key is stored in the `X-NatsBridge-LargeBody-Id`.
    - TODO: support for response body re-use based on cache etags?


## Concept

- HTTP => NATS => HTTP should functionally emit the same requests (and responses)
- Big Request bodies should be stored to JetStream
  - for transfer encoding chunked
- (NATS -> HTTP) Big response bodies should be stored to JetStream
  - for transfer encoding chunked
  - for unknown response sizes
  - TODO: Re-use same cache entries if etags match?
- TODO: is request / response streaming necessary???
- All HTTP headers are passed through to NATS without modification
- All Nats headers except "X-NatsBridge-...." are passed to HTTP without modification
  - X-NatsBridge-Method
  - X-NatsBridge-JSBodyId
- allow multiple NATS servers

## What's Next?
While this is currently functional and useful as is, here are the things I'd like to add next:

- [ ] Add more examples in the /examples directory
- [ ] Add Validation for all caddy modules in this package
- [ ] Add godoc comments
- [ ] Support mapping nats headers and http headers for upstream and downstream
- [ ] Customizable error handling
- [ ] Jetstream support (for all that persistence babyyyy)
- [ ] nats KV for storage
