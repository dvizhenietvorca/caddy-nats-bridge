package caddy_nats_bridge

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/dvizhenietvorca/caddy-nats-bridge/body_jetstream"
	"github.com/dvizhenietvorca/caddy-nats-bridge/logoutput"
	"github.com/dvizhenietvorca/caddy-nats-bridge/natsbridge"
	"github.com/dvizhenietvorca/caddy-nats-bridge/publish"
	"github.com/dvizhenietvorca/caddy-nats-bridge/request"
	"github.com/dvizhenietvorca/caddy-nats-bridge/subscribe"
)

func init() {
	caddy.RegisterModule(natsbridge.NatsBridgeApp{})
	httpcaddyfile.RegisterGlobalOption("nats", natsbridge.ParseGobalNatsOption)
	caddy.RegisterModule(subscribe.Subscribe{})

	caddy.RegisterModule(publish.Publish{})
	httpcaddyfile.RegisterHandlerDirective("nats_publish", publish.ParsePublishHandler)

	caddy.RegisterModule(request.Request{})
	httpcaddyfile.RegisterHandlerDirective("nats_request", request.ParseRequestHandler)

	// store request body to Jetstream
	caddy.RegisterModule(body_jetstream.StoreBodyToJetStream{})
	httpcaddyfile.RegisterHandlerDirective("store_body_to_jetstream", body_jetstream.ParseStoreBodyToJetstream)

	// logging output to NATS
	caddy.RegisterModule(logoutput.LogOutput{})
}
