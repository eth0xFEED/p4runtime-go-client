package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	p4_v1 "github.com/p4lang/p4runtime/go/p4/v1"

	"github.com/eth0xFEED/p4runtime-go-client/pkg/client"
	"github.com/eth0xFEED/p4runtime-go-client/pkg/signals"
)

const (
	defaultDeviceID = 1
)

var (
	defaultAddr = fmt.Sprintf("127.0.0.1:%d", client.P4RuntimePort)
)

func handleStreamMessages(p4RtC *client.Client, messageCh <-chan *p4_v1.StreamMessageResponse) {
	for message := range messageCh {
		switch message.Update.(type) {
		case *p4_v1.StreamMessageResponse_Packet:
			log.Debugf("Received PacketIn")
		case *p4_v1.StreamMessageResponse_Digest:
			log.Debugf("Received DigestList")
			// if err := learnMacs(p4RtC, m.Digest); err != nil {
			// 	log.Errorf("Error when learning MACs: %v", err)
			// }
		case *p4_v1.StreamMessageResponse_IdleTimeoutNotification:
			log.Debugf("Received IdleTimeoutNotification")
			// forgetEntries(p4RtC, m.IdleTimeoutNotification)
		case *p4_v1.StreamMessageResponse_Error:
			log.Errorf("Received StreamError")
		default:
			log.Errorf("Received unknown stream message")
		}
	}
}

func main() {
	var addr string
	flag.StringVar(&addr, "addr", defaultAddr, "P4Runtime server socket")
	var deviceID uint64
	flag.Uint64Var(&deviceID, "device-id", defaultDeviceID, "Device id")
	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose mode with debug log messages")
	var binPath string
	flag.StringVar(&binPath, "bin", "", "Path to P4 bin (not needed for bmv2 simple_switch_grpc)")
	var Xp4infoPath string
	flag.StringVar(&Xp4infoPath, "Xp4info", "", "Path to Xp4info (not needed for bmv2 simple_switch_grpc)")
	var caPath string
	flag.StringVar(&caPath, "caPath", "", "Path to ca certificate")

	flag.Parse()

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	binBytes, err := ioutil.ReadFile(binPath)
	if err != nil {
		log.Fatalf("Error when reading binary config from '%s': %v", binPath, err)
	}

	Xp4infoBytes, err := ioutil.ReadFile(Xp4infoPath)
	if err != nil {
		log.Fatalf("Error when reading Xp4info text file '%s': %v", Xp4infoPath, err)
	}

	opts := []grpc.DialOption{}
	log.Infof("Connecting to server at %s", addr)
	if caPath != "" {

		creds, err := loadTLSCredentials(caPath)
		if err != nil {
			log.Fatalf("Cannot load cert : %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		log.Fatalf("Cannot connect to server: %v", err)
	}
	defer conn.Close()

	c := p4_v1.NewP4RuntimeClient(conn)
	resp, err := c.Capabilities(context.Background(), &p4_v1.CapabilitiesRequest{})
	if err != nil {
		log.Fatalf("Error in Capabilities RPC: %v", err)
	}
	log.Infof("P4Runtime server version is %s", resp.P4RuntimeApiVersion)

	stopCh := signals.RegisterSignalHandlers()

	electionID := p4_v1.Uint128{High: 0, Low: 1}

	p4RtC := client.NewClient(c, deviceID, electionID)
	arbitrationCh := make(chan bool)
	messageCh := make(chan *p4_v1.StreamMessageResponse, 1000)
	defer close(messageCh)
	go p4RtC.Run(stopCh, arbitrationCh, messageCh)

	waitCh := make(chan struct{})

	go func() {
		sent := false
		for isPrimary := range arbitrationCh {
			if isPrimary {
				log.Infof("We are the primary client!")
				if !sent {
					waitCh <- struct{}{}
					sent = true
				}
			} else {
				log.Infof("We are not the primary client!")
			}
		}
	}()

	// it would also be safe to spawn multiple goroutines to handle messages from the channel
	go handleStreamMessages(p4RtC, messageCh)

	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	select {
	case <-ctx.Done():
		log.Fatalf("Could not become the primary client within %v", timeout)
	case <-waitCh:
	}

	log.Info("Setting forwarding pipe")
	if _, err := p4RtC.SetFwdPipeFromBytes(binBytes, Xp4infoBytes, 0); err != nil {
		log.Fatalf("Error when setting forwarding pipe: %v", err)
	}

	// log.Info("Do Ctrl-C to quit")
	// <-stopCh
	// log.Info("Stopping client")
}
func loadTLSCredentials(caPath string) (credentials.TransportCredentials, error) {
	return credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	}), nil
	// creds, err := credentials.NewClientTLSFromFile(caPath, "")
	// return creds, err
}
