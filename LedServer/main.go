package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "GreysonsLEDs"

var chatRoom *ChatRoom

const WebServer = "0.0.0.0:80"

const PubSubListeningAddr = "/ip4/0.0.0.0/tcp/0"

func main() {
	// parse some flags to set our nickname and the room to join
	deviceFlag := flag.String("device", "WebServer", "Device name to use. Will be \"Default\" if empty")
	roomFlag := flag.String("room", "LivingRoom", "name of the room to join")
	flag.Parse()

	ctx := context.Background()

	// create a new libp2p Host that listens on a random TCP port
	h, err := libp2p.New(libp2p.ListenAddrStrings(PubSubListeningAddr))
	if err != nil {
		panic(err)
	}

	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}

	// setup local mDNS discovery
	if err := setupDiscovery(h); err != nil {
		panic(err)
	}

	// use the nickname from the cli flag, or a default if blank
	nick := ""
	ip := strings.Split(h.Addrs()[0].String(), "/")[2]
	if len(*deviceFlag) != 0 {
		nick += *deviceFlag + "-" + shortID(h.ID()) + "@" + ip
	} else {
		nick += defaultNick(h.ID()) + "@" + ip
	}
	// join the room from the cli flag, or the flag default
	room := *roomFlag

	// join the chat room
	chatRoom, err = JoinChatRoom(ctx, ps, h.ID(), nick, room)
	if err != nil {
		panic(err)
	}
	addrs := ""
	for _, addr := range h.Addrs() {
		addrs += addr.String() + " "
	}
	http.HandleFunc("/", httpHandler)
	http.HandleFunc("/favicon.ico", httpIconHandler)
	http.HandleFunc("/setColor", httpSetColorHandler)
	http.HandleFunc("/getPeers", httpGetPeersHandler)
	fmt.Println("Listening on", WebServer)
	log.Fatal(http.ListenAndServe(WebServer, nil))
}

// defaultNick generates a nickname based on the $USER environment variable and
// the last 8 chars of a peer ID.
func defaultNick(p peer.ID) string {
	return fmt.Sprintf("%s-%s", "Default", shortID(p))
}

// shortID returns the last 8 chars of a base58-encoded peer id.
func shortID(p peer.ID) string {
	pretty := p.Pretty()
	return pretty[len(pretty)-8:]
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also~`	`
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	//fmt.Printf("discovered new peer %s\n", pi.ID.Pretty())
	n.h.Connect(context.Background(), pi)
	//if err != nil {
	//fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	//}
}

type Page struct {
	Title string
	Body  []byte
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	rawPage, err := ioutil.ReadFile("colorPickerPage.html")
	if err != nil {
		fmt.Printf("error reading html colorpicker page: %v", err)
	}
	fmt.Fprint(w, string(rawPage))
	fmt.Println("Connection from", r.RemoteAddr, "Request:", r.URL.Path)
}

func httpIconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
	fmt.Println("Connection from", r.RemoteAddr, "Request:", r.URL.Path)
}

func httpGetPeersHandler(w http.ResponseWriter, r *http.Request) {
	var shortPeers []string
	for _, peer := range chatRoom.ListPeers() {
		shortPeers = append(shortPeers, shortID(peer))
	}
	fmt.Fprint(w, "Self: "+chatRoom.nick+"\n", shortPeers)
	fmt.Println("Connection from", r.RemoteAddr, "Request:", r.URL.Path)
}

func httpSetColorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Println("Connection from", r.RemoteAddr, "Request:", r.URL.Path)
		fmt.Println(string(data))
		chatRoom.Publish(string(data))
	}
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host) error {
	// setup mDNS discovery
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}
