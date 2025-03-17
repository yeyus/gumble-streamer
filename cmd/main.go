package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"

	"layeh.com/gumble/gumble"
	"layeh.com/gumble/gumbleffmpeg"
	"layeh.com/gumble/gumbleutil"
)

type StreamerState int

const (
	StreamerStateDisconnected StreamerState = iota
	StreamerStateConnecting
	StreamerStateConnected
	StreamerStateIdle
	StreamerStateTalking
)

type Streamer struct {
	State  StreamerState
	Config *gumble.Config
	Client *gumble.Client

	Address   string
	TLSConfig *tls.Config

	Room string

	StreamAddress string
	Stream        *gumbleffmpeg.Stream

	WaitGroup *sync.WaitGroup
}

func NewStreamer(address string, room string, streamAddress string, config *gumble.Config, tlsConfig *tls.Config) *Streamer {
	return &Streamer{
		State:         StreamerStateDisconnected,
		Config:        config,
		TLSConfig:     tlsConfig,
		Address:       address,
		Room:          room,
		StreamAddress: streamAddress,
		WaitGroup:     new(sync.WaitGroup),
	}
}

func (s *Streamer) Connect() error {
	if s.State != StreamerStateDisconnected {
		return nil
	}

	s.State = StreamerStateConnecting
	s.WaitGroup.Add(1)

	s.Config.Attach(gumbleutil.Listener{
		Connect:       s.onConnect,
		Disconnect:    s.onDisconnect,
		ChannelChange: s.onChannelChange,
	})

	client, err := gumble.DialWithDialer(new(net.Dialer), s.Address, s.Config, s.TLSConfig)
	if err != nil {
		s.State = StreamerStateDisconnected
		fmt.Printf("[streamer:connect] error while connecting %v\n", err)
		return err
	}

	s.Client = client
	return nil
}

func (s *Streamer) onConnect(e *gumble.ConnectEvent) {
	s.State = StreamerStateConnected
	fmt.Printf("[streamer] connected to %s", s.Address)

	targetChannel := e.Client.Channels.Find(s.Room)
	if targetChannel == nil {
		fmt.Printf("[streamer] could not find channel %s, aborting\n", s.Room)
		e.Client.Disconnect()
		return
	}

	fmt.Printf("[streamer] moving to %s\n", targetChannel.Name)
	e.Client.Self.Move(targetChannel)
}

func (s *Streamer) onDisconnect(e *gumble.DisconnectEvent) {
	defer s.WaitGroup.Done()

	s.State = StreamerStateDisconnected
	fmt.Printf("[streamer] disconnected from %s\n", s.Address)
}

func (s *Streamer) onChannelChange(e *gumble.ChannelChangeEvent) {
	if len(e.Channel.Users) > 1 {
		s.State = StreamerStateIdle
	} else {
		s.State = StreamerStateConnected
	}
}

func main() {
	// Command line flags
	server := flag.String("server", "10.1.200.5:64738", "the server to connect to")
	username := flag.String("username", "", "the username of the client")
	password := flag.String("password", "", "the password of the server")
	insecure := flag.Bool("insecure", false, "skip server certificate verification")
	certificate := flag.String("certificate", "", "PEM encoded certificate and private key")

	room := flag.String("room", "", "The Room name where the streamer shall enter")
	stream := flag.String("stream", "", "The Stream to pipe into the room audio")

	flag.Parse()

	tlsConfig := &tls.Config{}
	if *insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	if *certificate != "" {
		cert, err := tls.LoadX509KeyPair(*certificate, *certificate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}

	config := gumble.NewConfig()
	config.Username = *username
	config.Password = *password

	streamer := NewStreamer(*server, *room, *stream, config, tlsConfig)

	streamer.Connect()

	streamer.WaitGroup.Wait()
}
