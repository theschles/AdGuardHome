package dhcpsvc

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/google/gopacket/layers"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// Config is the configuration for the DHCP service.
type Config struct {
	// Interfaces stores configurations of DHCP server specific for the network
	// interface identified by its name.
	Interfaces map[string]*InterfaceConfig

	// LocalDomainName is the top-level domain name to use for resolving DHCP
	// clients' hostnames.
	LocalDomainName string

	// ICMPTimeout is the timeout for checking another DHCP server's presence.
	ICMPTimeout time.Duration

	// Enabled is the state of the service, whether it is enabled or not.
	Enabled bool
}

// InterfaceConfig is the configuration of a single DHCP interface.
type InterfaceConfig struct {
	// IPv4 is the configuration of DHCP protocol for IPv4.
	IPv4 *IPv4Config

	// IPv6 is the configuration of DHCP protocol for IPv6.
	IPv6 *IPv6Config
}

// IPv4Config is the interface-specific configuration for DHCPv4.
type IPv4Config struct {
	// GatewayIP is the IPv4 address of the network's gateway.  It is used as
	// the default gateway for DHCP clients and also used in calculating the
	// network-specific broadcast address.
	GatewayIP netip.Addr

	// SubnetMask is the IPv4 subnet mask of the network.  It should be a valid
	// IPv4 CIDR (i.e. all 1s followed by all 0s).
	SubnetMask netip.Addr

	// RangeStart is the first address in the range to assign to DHCP clients.
	RangeStart netip.Addr

	// RangeEnd is the last address in the range to assign to DHCP clients.
	RangeEnd netip.Addr

	// Options is the list of DHCP options to send to DHCP clients.
	Options layers.DHCPOptions

	// LeaseDuration is the TTL of a DHCP lease.
	LeaseDuration time.Duration

	// Enabled is the state of the DHCPv4 service, whether it is enabled or not
	// on the specific interface.
	Enabled bool
}

// IPv6Config is the interface-specific configuration for DHCPv6.
type IPv6Config struct {
	// RangeStart is the first address in the range to assign to DHCP clients.
	RangeStart netip.Addr

	// Options is the list of DHCP options to send to DHCP clients.
	Options layers.DHCPOptions

	// LeaseDuration is the TTL of a DHCP lease.
	LeaseDuration time.Duration

	// RASlaacOnly defines whether the DHCP clients should only use SLAAC for
	// address assignment.
	RASLAACOnly bool

	// RAAllowSlaac defines whether the DHCP clients may use SLAAC for address
	// assignment.
	RAAllowSLAAC bool

	// Enabled is the state of the DHCPv6 service, whether it is enabled or not
	// on the specific interface.
	Enabled bool
}

// TODO(e.burkov):  !! doc
const ErrNilConfig errors.Error = "config is nil"

func (conf *Config) Validate() (err error) {
	switch {
	case conf == nil:
		return ErrNilConfig
	case !conf.Enabled:
		return nil
	case conf.ICMPTimeout < 0:
		return fmt.Errorf("icmp timeout %s must be non-negative", conf.ICMPTimeout)
	}

	err = netutil.ValidateDomainName(conf.LocalDomainName)
	if err != nil {
		// Don't wrap the error since it's informative enough as is.
		return err
	}

	ifaces := maps.Keys(conf.Interfaces)
	slices.Sort(ifaces)

	return errors.Join(
		errors.Annotate(conf.validateV4(ifaces), "validating v4: %w"),
		errors.Annotate(conf.validateV6(ifaces), "validating v6: %w"),
	)
}

func (conf *Config) validateV4(ifaces []string) (err error) {
	for _, iface := range ifaces {
		ifaceConf := conf.Interfaces[iface]
		if ifaceConf == nil {
			return ErrNilConfig
		}

		v4Conf := ifaceConf.IPv4
		switch {
		case !v4Conf.Enabled:
			continue
		case !v4Conf.GatewayIP.Is4():
			return fmt.Errorf("interface %q: gateway ip should be a valid ipv4", iface)
		case !v4Conf.SubnetMask.Is4():
			return fmt.Errorf("interface %q: subnet mask should be a valid ipv4 cidr", iface)
		case !v4Conf.RangeStart.Is4():
			return fmt.Errorf("interface %q: range start should be a valid ipv4", iface)
		case !v4Conf.RangeEnd.Is4():
			return fmt.Errorf("interface %q: range end should be a valid ipv4", iface)
		}

		c.ipRange, err = newIPRange(rangeStart.AsSlice(), rangeEnd.AsSlice())
		if err != nil {
			// Don't wrap the error since it's informative enough as is and there is
			// an annotation deferred already.
			return err
		}
	}

	return nil
}

func (conf *Config) validateV6(ifaces []string) (err error) {
	return nil
}
