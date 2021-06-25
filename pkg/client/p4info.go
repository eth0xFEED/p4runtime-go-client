package client

import (
	p4_config_v1 "github.com/p4lang/p4runtime/go/p4/config/v1"
)

const invalidID = 0

func (c *Client) tableId(name string) uint32 {
	if c.Xp4info == nil {
		return invalidID
	}
	for _, table := range c.Xp4info.Tables {
		if table.Preamble.Name == name {
			return table.Preamble.Id
		}
	}
	return invalidID
}

func (c *Client) actionProfileId(name string) uint32 {
	if c.Xp4info == nil {
		return invalidID
	}
	for _, ap := range c.Xp4info.ActionProfiles {
		if ap.Preamble.Name == name {
			return ap.Preamble.Id
		}
	}
	return invalidID

}
func (c *Client) actionId(name string) uint32 {
	if c.Xp4info == nil {
		return invalidID
	}
	for _, action := range c.Xp4info.Actions {
		if action.Preamble.Name == name {
			return action.Preamble.Id
		}
	}
	return invalidID
}

func (c *Client) digestId(name string) uint32 {
	if c.Xp4info == nil {
		return invalidID
	}
	for _, digest := range c.Xp4info.Digests {
		if digest.Preamble.Name == name {
			return digest.Preamble.Id
		}
	}
	return invalidID
}

func (c *Client) findCounter(name string) *p4_config_v1.Counter {
	if c.Xp4info == nil {
		return nil
	}
	for _, counter := range c.Xp4info.Counters {
		if counter.Preamble.Name == name {
			return counter
		}
	}
	return nil
}

func (c *Client) counterId(name string) uint32 {
	counter := c.findCounter(name)
	if counter == nil {
		return invalidID
	}
	return counter.Preamble.Id
}
