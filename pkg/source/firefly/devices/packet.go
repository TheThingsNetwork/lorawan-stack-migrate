package devices

import (
	"encoding/json"
	"io"

	"go.thethings.network/lorawan-stack-migrate/pkg/source/firefly/api"
)

type Packet struct {
	FCnt int `json:"fcnt"`
}

type JSONPacket struct {
	Packet Packet
}

func packetFromRequestBody(r io.ReadCloser) (*Packet, error) {
	defer r.Close()
	var packet JSONPacket
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&packet); err != nil {
		return nil, err
	}
	return &packet.Packet, nil
}

func GetLastPacket() (*Packet, error) {
	resp, err := api.GetLastPacket()
	if err != nil {
		return nil, err
	}
	return packetFromRequestBody(resp.Body)
}

type JSONPackets struct {
	Packets []Packet
}

func packetListFromRequestBody(r io.ReadCloser) ([]Packet, error) {
	defer r.Close()
	var packets JSONPackets
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&packets); err != nil {
		return nil, err
	}
	return packets.Packets, nil
}

func GetPacketList() ([]Packet, error) {
	resp, err := api.GetPacketList()
	if err != nil {
		return nil, err
	}
	return packetListFromRequestBody(resp.Body)
}
