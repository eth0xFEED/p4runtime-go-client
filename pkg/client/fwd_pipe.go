package client

import (
	"context"
	"fmt"
	"io/ioutil"

	//nolint:staticcheck // SA1019 To be resolved later
	//lint:ignore SA1019 This line added for support golint version of VSC
	"github.com/golang/protobuf/proto"

	p4_config_v1 "github.com/p4lang/p4runtime/go/p4/config/v1"
	p4_v1 "github.com/p4lang/p4runtime/go/p4/v1"
)

type FwdPipeConfig struct {
	Xp4info        *p4_config_v1.P4Info
	P4DeviceConfig []byte
	Cookie         uint64
}

func (c *Client) SetFwdPipeFromBytes(binBytes, p4infoBytes []byte, cookie uint64) (*FwdPipeConfig, error) {
	Xp4info := &p4_config_v1.P4Info{}
	if err := proto.UnmarshalText(string(p4infoBytes), Xp4info); err != nil {
		return nil, fmt.Errorf("failed to decode Xp4info Protobuf message: %v", err)
	}
	config := &p4_v1.ForwardingPipelineConfig{
		P4Info:         Xp4info,
		P4DeviceConfig: binBytes,
		Cookie: &p4_v1.ForwardingPipelineConfig_Cookie{
			Cookie: cookie,
		},
	}

	req := &p4_v1.SetForwardingPipelineConfigRequest{
		DeviceId:   c.deviceID,
		ElectionId: &c.electionID,
		Action:     p4_v1.SetForwardingPipelineConfigRequest_VERIFY_AND_COMMIT,
		Config:     config,
	}
	_, err := c.SetForwardingPipelineConfig(context.Background(), req)
	if err == nil {
		c.Xp4info = Xp4info
		return &FwdPipeConfig{
			Xp4info:        Xp4info,
			P4DeviceConfig: binBytes,
			Cookie:         cookie,
		}, nil
	}

	return nil, err
}

func (c *Client) SetFwdPipe(binPath string, Xp4infoPath string, cookie uint64) (*FwdPipeConfig, error) {
	binBytes, err := ioutil.ReadFile(binPath)
	if err != nil {
		return nil, fmt.Errorf("error when reading binary device config: %v", err)
	}
	Xp4infoBytes, err := ioutil.ReadFile(Xp4infoPath)
	if err != nil {
		return nil, fmt.Errorf("error when reading Xp4info text file: %v", err)
	}
	return c.SetFwdPipeFromBytes(binBytes, Xp4infoBytes, cookie)
}

type GetFwdPipeResponseType int32

const (
	GetFwdPipeAll                   = GetFwdPipeResponseType(p4_v1.GetForwardingPipelineConfigRequest_ALL)
	GetFwdPipeCookieOnly            = GetFwdPipeResponseType(p4_v1.GetForwardingPipelineConfigRequest_COOKIE_ONLY)
	GetFwdPipeXp4infoAndCookie      = GetFwdPipeResponseType(p4_v1.GetForwardingPipelineConfigRequest_P4INFO_AND_COOKIE)
	GetFwdPipeDeviceConfigAndCookie = GetFwdPipeResponseType(p4_v1.GetForwardingPipelineConfigRequest_DEVICE_CONFIG_AND_COOKIE)
)

// GetFwdPipe retrieves the current pipeline config used in the remote switch.
//
// responseType is oneof:
//  GetFwdPipeAll, GetFwdPipeCookieOnly, GetFwdPipeXp4infoAndCookie, GetFwdPipeDeviceConfigAndCookie
// See https://p4.org/p4runtime/spec/v1.3.0/P4Runtime-Spec.html#sec-getforwardingpipelineconfig-rpc
func (c *Client) GetFwdPipe(responseType GetFwdPipeResponseType) (*FwdPipeConfig, error) {
	req := &p4_v1.GetForwardingPipelineConfigRequest{
		DeviceId:     c.deviceID,
		ResponseType: p4_v1.GetForwardingPipelineConfigRequest_ResponseType(responseType),
	}

	resp, err := c.GetForwardingPipelineConfig(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("error when retrieving forwardingpipeline config: %v", err)
	}

	config := resp.GetConfig()
	if config == nil {
		// pipeline doesn't have a config yet
		return nil, nil
	}

	var pipeConfig = &FwdPipeConfig{
		Xp4info:        config.GetP4Info(),
		P4DeviceConfig: config.GetP4DeviceConfig(),
	}
	if Cookie := config.GetCookie(); Cookie != nil {
		pipeConfig.Cookie = Cookie.GetCookie()
	}

	// save Xp4info for later use
	if pipeConfig.Xp4info != nil {
		c.Xp4info = pipeConfig.Xp4info
	}

	return pipeConfig, nil
}
