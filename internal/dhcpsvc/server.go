package dhcpsvc

import (
	"fmt"
	"net"
	"net/netip"
	"sync/atomic"
	"time"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// iface4 is a DHCP interface for IPv4 address family.
type iface4 struct {
	// gateway is the IP address of the network gateway.
	gateway netip.Addr

	// subnet is the network subnet.
	//
	// TODO(e.burkov):  Make netip.Addr?
	subnet netip.Prefix

	// addrSpace is the IPv4 address space allocated for leasing.
	addrSpace *ipRange

	// name is the name of the interface.
	name string

	// TODO(e.burkov):  Add options.

	// leaseTTL is the time-to-live of dynamic leases on this interface.
	leaseTTL time.Duration
}

// newIface4 creates a new DHCP interface for IPv4 address family with the given
// configuration.  It returns an error if the given configuration can't be used.
func newIface4(name string, conf *IPv4Config) (i *iface4, err error) {
	if !conf.Enabled {
		return nil, nil
	}

	maskLen, _ := net.IPMask(conf.SubnetMask.AsSlice()).Size()
	subnet := netip.PrefixFrom(conf.GatewayIP, maskLen)

	switch {
	case !subnet.Contains(conf.RangeStart):
		return nil, fmt.Errorf("range start %s is not within %s", conf.RangeStart, subnet)
	case !subnet.Contains(conf.RangeEnd):
		return nil, fmt.Errorf("range end %s is not within %s", conf.RangeEnd, subnet)
	}

	addrSpace, err := newIPRange(conf.RangeStart, conf.RangeEnd)
	if err != nil {
		return nil, err
	} else if addrSpace.contains(conf.GatewayIP) {
		return nil, fmt.Errorf("gateway ip %s in the ip range %s", conf.GatewayIP, addrSpace)
	}

	return &iface4{
		name:      name,
		gateway:   conf.GatewayIP,
		subnet:    subnet,
		addrSpace: addrSpace,
		leaseTTL:  conf.LeaseDuration,
	}, nil
}

// iface6 is a DHCP interface for IPv6 address family.
//
// TODO(e.burkov):  Add options.
type iface6 struct {
	// rangeStart is the first IP address in the range.
	rangeStart netip.Addr

	// name is the name of the interface.
	name string

	// leaseTTL is the time-to-live of dynamic leases on this interface.
	leaseTTL time.Duration

	// raSLAACOnly defines if DHCP should send ICMPv6.RA packets without MO
	// flags.
	raSLAACOnly bool

	// raAllowSLAAC defines if DHCP should send ICMPv6.RA packets with MO flags.
	raAllowSLAAC bool
}

// newIface6 creates a new DHCP interface for IPv6 address family with the given
// configuration.
//
// TODO(e.burkov):  Validate properly.
func newIface6(name string, conf *IPv6Config) (i *iface6) {
	if !conf.Enabled {
		return nil
	}

	return &iface6{
		name:         name,
		rangeStart:   conf.RangeStart,
		leaseTTL:     conf.LeaseDuration,
		raSLAACOnly:  conf.RASLAACOnly,
		raAllowSLAAC: conf.RAAllowSLAAC,
	}
}

// DHCPServer is a DHCP server for both IPv4 and IPv6 address families.
type DHCPServer struct {
	// enabled indicates whether the DHCP server is enabled and can provide
	// information about its clients.
	enabled *atomic.Bool

	// interfaces4 is the set of IPv4 interfaces sorted by interface name.
	interfaces4 []*iface4

	// interfaces6 is the set of IPv6 interfaces sorted by interface name.
	interfaces6 []*iface6
}

// New creates a new DHCP server with the given configuration.  It returns an
// error if the given configuration can't be used.
func New(conf *Config) (srv *DHCPServer, err error) {
	if err = conf.Validate(); err != nil {
		// Don't wrap the error since it's informative enough as is.
		return nil, err
	} else if !conf.Enabled {
		// TODO(e.burkov):  !! return Empty?
		return nil, nil
	}

	ifaces4 := make([]*iface4, len(conf.Interfaces))
	ifaces6 := make([]*iface6, len(conf.Interfaces))

	ifaceNames := maps.Keys(conf.Interfaces)
	slices.Sort(ifaceNames)

	var i4 *iface4
	var i6 *iface6

	for _, ifaceName := range ifaceNames {
		iface := conf.Interfaces[ifaceName]

		i4, err = newIface4(ifaceName, iface.IPv4)
		if err != nil {
			return nil, fmt.Errorf("interface %q: ipv4: %w", ifaceName, err)
		} else if i4 != nil {
			ifaces4 = append(ifaces4, i4)
		}

		i6 = newIface6(ifaceName, iface.IPv6)
		if i6 != nil {
			ifaces6 = append(ifaces6, i6)
		}
	}

	enabled := &atomic.Bool{}
	enabled.Store(conf.Enabled)

	return &DHCPServer{
		enabled:     enabled,
		interfaces4: ifaces4,
		interfaces6: ifaces6,
	}, nil
}
