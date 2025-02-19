package client

import (
	p4_v1 "github.com/p4lang/p4runtime/go/p4/v1"
)

func (c *Client) InsertMulticastGroupEntry(mge *p4_v1.MulticastGroupEntry) error {

	preEntry := &p4_v1.PacketReplicationEngineEntry{
		Type: &p4_v1.PacketReplicationEngineEntry_MulticastGroupEntry{
			MulticastGroupEntry: mge,
		},
	}

	updateType := p4_v1.Update_INSERT
	update := &p4_v1.Update{
		Type: updateType,
		Entity: &p4_v1.Entity{
			Entity: &p4_v1.Entity_PacketReplicationEngineEntry{
				PacketReplicationEngineEntry: preEntry,
			},
		},
	}

	return c.WriteUpdate(update)
}
func (c *Client) InsertMulticastGroup(mgid uint32, ports []uint32) error {
	entry := &p4_v1.MulticastGroupEntry{
		MulticastGroupId: mgid,
	}
	for idx, port := range ports {
		replica := &p4_v1.Replica{
			EgressPort: port,
			Instance:   uint32(idx + 1),
		}
		entry.Replicas = append(entry.Replicas, replica)
	}

	preEntry := &p4_v1.PacketReplicationEngineEntry{
		Type: &p4_v1.PacketReplicationEngineEntry_MulticastGroupEntry{
			MulticastGroupEntry: entry,
		},
	}

	updateType := p4_v1.Update_INSERT
	update := &p4_v1.Update{
		Type: updateType,
		Entity: &p4_v1.Entity{
			Entity: &p4_v1.Entity_PacketReplicationEngineEntry{
				PacketReplicationEngineEntry: preEntry,
			},
		},
	}

	return c.WriteUpdate(update)
}

func (c *Client) DeleteMulticastGroup(mgid uint32) error {
	entry := &p4_v1.MulticastGroupEntry{
		MulticastGroupId: mgid,
	}

	preEntry := &p4_v1.PacketReplicationEngineEntry{
		Type: &p4_v1.PacketReplicationEngineEntry_MulticastGroupEntry{
			MulticastGroupEntry: entry,
		},
	}

	updateType := p4_v1.Update_DELETE
	update := &p4_v1.Update{
		Type: updateType,
		Entity: &p4_v1.Entity{
			Entity: &p4_v1.Entity_PacketReplicationEngineEntry{
				PacketReplicationEngineEntry: preEntry,
			},
		},
	}

	return c.WriteUpdate(update)
}

func (c *Client) ReadMulticastGroup(mgid uint32) (*p4_v1.Entity, error) {

	entity := &p4_v1.Entity{
		Entity: &p4_v1.Entity_PacketReplicationEngineEntry{
			PacketReplicationEngineEntry: &p4_v1.PacketReplicationEngineEntry{
				Type: &p4_v1.PacketReplicationEngineEntry_MulticastGroupEntry{
					MulticastGroupEntry: &p4_v1.MulticastGroupEntry{
						MulticastGroupId: mgid,
					},
				},
			},
		},
	}

	return c.ReadEntitySingle(entity)
}
