package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	multiaddr "github.com/multiformats/go-multiaddr"
)

// handleStream handles incoming streams from other peers
func handleStream(stream network.Stream) {
	log.Println("Received a new stream!")
	defer stream.Close()

	reader := bufio.NewReader(stream)
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading from stream:", err)
			return
		}

		fmt.Printf("Received message: %s", str)
	}
}

// StartNode starts a new libp2p node
func StartNode(listenAddr string) host.Host {
	// Create a new libp2p node
	h, err := libp2p.New(libp2p.ListenAddrStrings(listenAddr))
	if err != nil {
		log.Fatalln("Failed to create libp2p host:", err)
	}

	// Set a handler for incoming streams
	h.SetStreamHandler("/p2p/1.0.0", handleStream)

	// Get the peer information of the node
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%s/p2p/%s", listenAddr, h.ID().String()))
	fmt.Println("Node is listening on:", hostAddr)

	return h
}

// ConnectToPeer connects to another peer at the given address
func ConnectToPeer(h host.Host, peerAddr string) {
	// Convert the string address to a Multiaddr
	addr, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		log.Fatalln("Invalid peer address:", err)
	}

	// Extract peer information from the address
	peerinfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		log.Fatalln("Failed to get peer info:", err)
	}

	// Add the peer to the peer store and attempt to connect
	h.Peerstore().AddAddrs(peerinfo.ID, peerinfo.Addrs, peerstore.PermanentAddrTTL)
	if err := h.Connect(context.Background(), *peerinfo); err != nil {
		log.Fatalln("Failed to connect to peer:", err)
	}

	fmt.Println("Connected to peer:", peerinfo.ID.String())
}

// SendMessage sends a message to the connected peer
func SendMessage(h host.Host, peerIDStr string, message string) {
	// Parse the peer ID string
	peerID, err := peer.Decode(peerIDStr)
	if err != nil {
		log.Fatalln("Invalid peer ID:", err)
	}

	// Open a stream to the peer
	stream, err := h.NewStream(context.Background(), peerID, "/p2p/1.0.0")
	if err != nil {
		log.Fatalln("Failed to open stream:", err)
	}
	defer stream.Close()

	// Send the message to the peer
	_, err = stream.Write([]byte(message + "\n"))
	if err != nil {
		log.Fatalln("Failed to send message:", err)
	}

	fmt.Println("Message sent to", peerID.String())
}

func main() {
	// Command-line argument handling for the peer
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <listen-port>")
		return
	}

	// Start the node at the given port
	listenAddr := os.Args[1]
	node := StartNode(listenAddr)

	// Scanner to handle user input
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\nEnter command (connect <peer-addr>, send <peer-id> <message>, exit):")
		scanner.Scan()
		input := scanner.Text()
		args := strings.Split(input, " ")

		switch args[0] {
		case "connect":
			if len(args) != 2 {
				fmt.Println("Usage: connect <peer-addr>")
				continue
			}
			ConnectToPeer(node, args[1])

		case "send":
			if len(args) < 3 {
				fmt.Println("Usage: send <peer-id> <message>")
				continue
			}
			peerID := args[1]
			message := strings.Join(args[2:], " ")
			SendMessage(node, peerID, message)

		case "exit":
			fmt.Println("Exiting...")
			return

		default:
			fmt.Println("Unknown command")
		}
	}
}
