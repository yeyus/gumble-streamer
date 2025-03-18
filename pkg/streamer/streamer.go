package streamer

import (
	"crypto/tls"
	"fmt"
	"maps"
	"net"
	"sync"

	"github.com/yeyus/gumble-streamer/pkg/ffmpegsource"
	"github.com/yeyus/gumble/gumble"
	"github.com/yeyus/gumble/gumbleffmpeg"
	"github.com/yeyus/gumble/gumbleutil"
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

	Room []string

	StreamAddress     string
	Stream            *gumbleffmpeg.Stream
	StreamHTTPHeaders *ffmpegsource.HTTPHeaders
	StreamExtraParams *ffmpegsource.Params

	WaitGroup *sync.WaitGroup
}

func NewStreamer(address string, room []string, streamAddress string, config *gumble.Config, tlsConfig *tls.Config, httpHeaders *ffmpegsource.HTTPHeaders, extraParams *ffmpegsource.Params) *Streamer {
	return &Streamer{
		State:             StreamerStateDisconnected,
		Config:            config,
		TLSConfig:         tlsConfig,
		Address:           address,
		Room:              room,
		StreamAddress:     streamAddress,
		StreamHTTPHeaders: httpHeaders,
		StreamExtraParams: extraParams,
		WaitGroup:         new(sync.WaitGroup),
	}
}

func (s *Streamer) Connect() error {
	if s.State != StreamerStateDisconnected {
		return nil
	}

	s.State = StreamerStateConnecting
	s.WaitGroup.Add(1)

	s.Config.Attach(gumbleutil.Listener{
		Connect:    s.onConnect,
		Disconnect: s.onDisconnect,
		UserChange: s.onUserChange,
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

	targetChannel := e.Client.Channels.Find(s.Room...)
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

func (s *Streamer) onUserChange(e *gumble.UserChangeEvent) {
	if e.Type == gumble.UserChangeChannel {
		// users connected to our channel
		numUsersInChannel := len(e.Client.Self.Channel.Users)
		fmt.Printf("[streamer:onUserChange] users in our channel %d!\n", numUsersInChannel)

		if numUsersInChannel > 1 && s.State == StreamerStateConnected {
			// someone arrived
			s.State = StreamerStateIdle
			fmt.Printf("[streamer:onChannelChange] someone is in the channel, new state is %v\n", s.State)
			s.StartStreaming()
		} else if numUsersInChannel <= 1 && s.State == StreamerStateIdle {
			// I'm alone here
			s.State = StreamerStateConnected
			fmt.Printf("[streamer:onChannelChange] everyone left the room, new state is %v\n", s.State)
			s.StopStreaming()
		}
	}
}

func (s *Streamer) StartStreaming() error {
	if s.Stream != nil {
		panic("[streamer:stream] already have a stream")
	}

	ffmpegSource := ffmpegsource.NewFFMPEGSource(s.StreamAddress)

	ffmpegSource.HTTPHeaders = maps.Clone(*s.StreamHTTPHeaders)
	copy(ffmpegSource.ExtraParams, *s.StreamExtraParams)

	fmt.Printf("[streamer:stream] starting stream of %s\n", s.StreamAddress)
	s.Stream = gumbleffmpeg.New(s.Client, ffmpegSource)
	if err := s.Stream.Play(); err != nil {
		fmt.Printf("[streamer:stream] error launching ffmpeg: %s\n", err)
		return err
	} else {
		fmt.Printf("Playing %s\n", s.StreamAddress)
	}

	return nil
}

func (s *Streamer) StopStreaming() {
	if s.Stream != nil {
		fmt.Printf("[streamer:stream] stopping stream of %s\n", s.StreamAddress)
		s.Stream.Stop()
		s.Stream = nil
	}
}
