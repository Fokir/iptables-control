//go:build linux

package nftables

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

const tableName = "system_control"

// Apply atomically replaces all NAT rules in the system_control table.
func (e *Engine) Apply(groups []NatGroup) error {
	conn, err := nftables.New()
	if err != nil {
		return fmt.Errorf("nftables connect: %w", err)
	}

	// Delete existing table if present (ignore error if not exists)
	conn.DelTable(&nftables.Table{Name: tableName, Family: nftables.TableFamilyIPv4})
	conn.Flush()

	// Create fresh table
	table := conn.AddTable(&nftables.Table{
		Name:   tableName,
		Family: nftables.TableFamilyIPv4,
	})

	prerouting := conn.AddChain(&nftables.Chain{
		Name:     "prerouting",
		Table:    table,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityNATDest,
	})

	postrouting := conn.AddChain(&nftables.Chain{
		Name:     "postrouting",
		Table:    table,
		Type:     nftables.ChainTypeNAT,
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityNATSource,
	})

	for _, group := range groups {
		if !group.Enabled {
			continue
		}

		destIP := net.ParseIP(group.DestinationIP).To4()
		targetIP := net.ParseIP(group.TargetIP).To4()
		snatIP := net.ParseIP(group.TargetReverseIP).To4()

		if destIP == nil || targetIP == nil || snatIP == nil {
			slog.Error("invalid IP in group, skipping", "group", group.Name)
			continue
		}

		for _, port := range group.Ports {
			type protoInfo struct {
				proto byte
				name  string
			}
			var protocols []protoInfo
			if port.ProtocolTCP {
				protocols = append(protocols, protoInfo{unix.IPPROTO_TCP, "tcp"})
			}
			if port.ProtocolUDP {
				protocols = append(protocols, protoInfo{unix.IPPROTO_UDP, "udp"})
			}

			for _, p := range protocols {
				conn.AddRule(&nftables.Rule{
					Table: table,
					Chain: prerouting,
					Exprs: buildDNATExprs(destIP, p.proto, uint16(port.ExternalPort), targetIP, uint16(port.InternalPort)),
				})

				conn.AddRule(&nftables.Rule{
					Table: table,
					Chain: postrouting,
					Exprs: buildSNATExprs(targetIP, p.proto, uint16(port.InternalPort), snatIP),
				})

				slog.Debug("added NAT rule",
					"group", group.Name,
					"proto", p.name,
					"dest", fmt.Sprintf("%s:%d", destIP, port.ExternalPort),
					"target", fmt.Sprintf("%s:%d", targetIP, port.InternalPort),
					"snat", snatIP.String(),
				)
			}
		}
	}

	if err := conn.Flush(); err != nil {
		return fmt.Errorf("nftables flush: %w", err)
	}

	slog.Info("nftables rules applied", "groups", len(groups))
	return nil
}

func buildDNATExprs(destIP net.IP, proto byte, dport uint16, targetIP net.IP, targetPort uint16) []expr.Any {
	return []expr.Any{
		&expr.Payload{DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 16, Len: 4},
		&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: destIP},
		&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
		&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: []byte{proto}},
		&expr.Payload{DestRegister: 1, Base: expr.PayloadBaseTransportHeader, Offset: 2, Len: 2},
		&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: binaryPort(dport)},
		&expr.Immediate{Register: 1, Data: targetIP},
		&expr.Immediate{Register: 2, Data: binaryPort(targetPort)},
		&expr.NAT{
			Type:        expr.NATTypeDestNAT,
			Family:      unix.NFPROTO_IPV4,
			RegAddrMin:  1,
			RegProtoMin: 2,
		},
	}
}

func buildSNATExprs(targetIP net.IP, proto byte, dport uint16, snatIP net.IP) []expr.Any {
	return []expr.Any{
		&expr.Payload{DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 16, Len: 4},
		&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: targetIP},
		&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
		&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: []byte{proto}},
		&expr.Payload{DestRegister: 1, Base: expr.PayloadBaseTransportHeader, Offset: 2, Len: 2},
		&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: binaryPort(dport)},
		&expr.Immediate{Register: 1, Data: snatIP},
		&expr.NAT{
			Type:       expr.NATTypeSourceNAT,
			Family:     unix.NFPROTO_IPV4,
			RegAddrMin: 1,
		},
	}
}

func binaryPort(port uint16) []byte {
	return []byte{byte(port >> 8), byte(port & 0xff)}
}
