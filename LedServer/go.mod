module github.com/libp2p/go-libp2p/examples/pubsub/LedServer

go 1.16

require (
	github.com/libp2p/go-libp2p v0.15.1
	github.com/libp2p/go-libp2p-core v0.11.0
	github.com/libp2p/go-libp2p-pubsub v0.5.3
)

// Ensure that examples always use the go-libp2p version in the same git checkout.
replace github.com/libp2p/go-libp2p => ../
