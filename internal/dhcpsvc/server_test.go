package dhcpsvc_test

import (
	"net/netip"
	"testing"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/dhcpsvc"
	"github.com/AdguardTeam/golibs/testutil"
)

func TestNew(t *testing.T) {
	const validLocalTLD = "local"

	validIPv4Conf := &dhcpsvc.IPv4Config{
		Enabled:       true,
		GatewayIP:     netip.MustParseAddr("192.168.0.1"),
		SubnetMask:    netip.MustParseAddr("255.255.255.0"),
		RangeStart:    netip.MustParseAddr("192.168.0.2"),
		RangeEnd:      netip.MustParseAddr("192.168.0.254"),
		LeaseDuration: 1 * time.Hour,
	}
	gwInRangeConf := &dhcpsvc.IPv4Config{
		Enabled:       true,
		GatewayIP:     netip.MustParseAddr("192.168.0.100"),
		SubnetMask:    netip.MustParseAddr("255.255.255.0"),
		RangeStart:    netip.MustParseAddr("192.168.0.1"),
		RangeEnd:      netip.MustParseAddr("192.168.0.254"),
		LeaseDuration: 1 * time.Hour,
	}
	badStartConf := &dhcpsvc.IPv4Config{
		Enabled:       true,
		GatewayIP:     netip.MustParseAddr("192.168.0.1"),
		SubnetMask:    netip.MustParseAddr("255.255.255.0"),
		RangeStart:    netip.MustParseAddr("127.0.0.1"),
		RangeEnd:      netip.MustParseAddr("192.168.0.254"),
		LeaseDuration: 1 * time.Hour,
	}

	validIPv6Conf := &dhcpsvc.IPv6Config{
		Enabled:       true,
		RangeStart:    netip.MustParseAddr("2001:db8::1"),
		LeaseDuration: 1 * time.Hour,
		RAAllowSLAAC:  true,
		RASLAACOnly:   true,
	}

	testCases := []struct {
		conf       *dhcpsvc.Config
		name       string
		wantErrMsg string
	}{{
		conf:       nil,
		name:       "nil_config",
		wantErrMsg: "config is nil",
	}, {
		conf: &dhcpsvc.Config{
			Enabled: false,
		},
		name:       "disabled",
		wantErrMsg: "",
	}, {
		conf: &dhcpsvc.Config{
			Enabled: true,
		},
		name:       "bad_local_tld",
		wantErrMsg: `bad domain name "": domain name is empty`,
	}, {
		conf: &dhcpsvc.Config{
			Enabled:         true,
			LocalDomainName: validLocalTLD,
			Interfaces:      nil,
		},
		name:       "no_interfaces",
		wantErrMsg: "no interfaces specified",
	}, {
		conf: &dhcpsvc.Config{
			Enabled:         true,
			LocalDomainName: validLocalTLD,
			Interfaces: map[string]*dhcpsvc.InterfaceConfig{
				"eth0": nil,
			},
		},
		name:       "nil_interface",
		wantErrMsg: `interface "eth0": config is nil`,
	}, {
		conf: &dhcpsvc.Config{
			Enabled:         true,
			LocalDomainName: validLocalTLD,
			Interfaces: map[string]*dhcpsvc.InterfaceConfig{
				"eth0": {
					IPv4: nil,
					IPv6: &dhcpsvc.IPv6Config{Enabled: false},
				},
			},
		},
		name:       "nil_ipv4",
		wantErrMsg: `interface "eth0": ipv4: config is nil`,
	}, {
		conf: &dhcpsvc.Config{
			Enabled:         true,
			LocalDomainName: validLocalTLD,
			Interfaces: map[string]*dhcpsvc.InterfaceConfig{
				"eth0": {
					IPv4: &dhcpsvc.IPv4Config{Enabled: false},
					IPv6: nil,
				},
			},
		},
		name:       "nil_ipv6",
		wantErrMsg: `interface "eth0": ipv6: config is nil`,
	}, {
		conf: &dhcpsvc.Config{
			Enabled:         true,
			LocalDomainName: validLocalTLD,
			Interfaces: map[string]*dhcpsvc.InterfaceConfig{
				"eth0": {
					IPv4: validIPv4Conf,
					IPv6: validIPv6Conf,
				},
			},
		},
		name:       "valid",
		wantErrMsg: "",
	}, {
		conf: &dhcpsvc.Config{
			Enabled:         true,
			LocalDomainName: validLocalTLD,
			Interfaces: map[string]*dhcpsvc.InterfaceConfig{
				"eth0": {
					IPv4: &dhcpsvc.IPv4Config{Enabled: false},
					IPv6: &dhcpsvc.IPv6Config{Enabled: false},
				},
			},
		},
		name:       "disabled_interfaces",
		wantErrMsg: "",
	}, {
		conf: &dhcpsvc.Config{
			Enabled:         true,
			LocalDomainName: validLocalTLD,
			Interfaces: map[string]*dhcpsvc.InterfaceConfig{
				"eth0": {
					IPv4: gwInRangeConf,
					IPv6: validIPv6Conf,
				},
			},
		},
		name: "gateway_within_range",
		wantErrMsg: `interface "eth0": ipv4: ` +
			`gateway ip 192.168.0.100 in the ip range 192.168.0.1-192.168.0.254`,
	}, {
		conf: &dhcpsvc.Config{
			Enabled:         true,
			LocalDomainName: validLocalTLD,
			Interfaces: map[string]*dhcpsvc.InterfaceConfig{
				"eth0": {
					IPv4: badStartConf,
					IPv6: validIPv6Conf,
				},
			},
		},
		name: "bad_start",
		wantErrMsg: `interface "eth0": ipv4: ` +
			`range start 127.0.0.1 is not within 192.168.0.1/24`,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := dhcpsvc.New(tc.conf)
			testutil.AssertErrorMsg(t, tc.wantErrMsg, err)
		})
	}
}
