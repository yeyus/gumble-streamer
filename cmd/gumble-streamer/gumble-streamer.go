package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yeyus/gumble-streamer/pkg/ffmpegsource"
	"github.com/yeyus/gumble-streamer/pkg/streamer"

	"github.com/yeyus/gumble/gumble"
	_ "github.com/yeyus/gumble/opus"
)

func main() {
	// Command line flags
	server := flag.String("server", "127.0.0.1:64738", "the server to connect to")
	username := flag.String("username", "", "the username of the client")
	password := flag.String("password", "", "the password of the server")
	insecure := flag.Bool("insecure", false, "skip server certificate verification")
	certificate := flag.String("certificate", "", "PEM encoded certificate and private key")

	room := flag.String("room", "", "The Room path separated by commas where the streamer shall enter")
	stream := flag.String("stream", "", "The Stream to pipe into the room audio")
	referer := flag.String("referer", "", "The referer sent to the streaming server")
	userAgent := flag.String("useragent", "", "The user agent sent to the streaming server")
	origin := flag.String("origin", "", "The origin sent to the streaming server")

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

	httpHeaders := make(ffmpegsource.HTTPHeaders)

	if len(*referer) > 0 {
		httpHeaders["Referer"] = *referer
	}

	if len(*userAgent) > 0 {
		httpHeaders["User-Agent"] = *userAgent
	}

	if len(*origin) > 0 {
		httpHeaders["Origin"] = *origin
	}

	extraParams := make(ffmpegsource.Params, 0)

	streamer := streamer.NewStreamer(*server, strings.Split(*room, ","), *stream, config, tlsConfig, &httpHeaders, &extraParams)

	streamer.Connect()

	streamer.WaitGroup.Wait()
}
