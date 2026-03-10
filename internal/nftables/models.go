package nftables

import "time"

type NatGroup struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Enabled         bool       `json:"enabled"`
	TargetIP        string     `json:"targetIp"`
	TargetReverseIP string     `json:"targetReverseIp"`
	DestinationIP   string     `json:"destinationIp"`
	Ports           []NatPort  `json:"ports"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type NatPort struct {
	ID           int64     `json:"id"`
	GroupID      int64     `json:"groupId"`
	ExternalPort int       `json:"externalPort"`
	InternalPort int       `json:"internalPort"`
	ProtocolTCP  bool      `json:"protocolTcp"`
	ProtocolUDP  bool      `json:"protocolUdp"`
	CreatedAt    time.Time `json:"createdAt"`
}

type CreateGroupRequest struct {
	Name            string           `json:"name"`
	Enabled         bool             `json:"enabled"`
	TargetIP        string           `json:"targetIp"`
	TargetReverseIP string           `json:"targetReverseIp"`
	DestinationIP   string           `json:"destinationIp"`
	Ports           []CreatePortRequest `json:"ports"`
}

type UpdateGroupRequest struct {
	Name            string           `json:"name"`
	Enabled         bool             `json:"enabled"`
	TargetIP        string           `json:"targetIp"`
	TargetReverseIP string           `json:"targetReverseIp"`
	DestinationIP   string           `json:"destinationIp"`
	Ports           []CreatePortRequest `json:"ports"`
}

type CreatePortRequest struct {
	ExternalPort int  `json:"externalPort"`
	InternalPort int  `json:"internalPort"`
	ProtocolTCP  bool `json:"protocolTcp"`
	ProtocolUDP  bool `json:"protocolUdp"`
}
