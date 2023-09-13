package dhcpsvc

import (
	"fmt"
	"math"
	"math/big"
	"net"
	"net/netip"

	"github.com/AdguardTeam/golibs/errors"
)

// ipRange is an inclusive range of IP addresses.  A nil range is a range that
// doesn't contain any IP addresses.
//
// It is safe for concurrent use.
//
// TODO(a.garipov): Perhaps create an optimized version with uint32 for IPv4
// ranges?  Or use one of uint128 packages?
type ipRange struct {
	start *big.Int
	end   *big.Int
}

// maxRangeLen is the maximum IP range length.  The bitsets used in servers only
// accept uints, which can have the size of 32 bit.
const maxRangeLen = math.MaxUint32

// newIPRange creates a new IP address range.  start must be less than end.  The
// resulting range must not be greater than maxRangeLen.
func newIPRange(start, end netip.Addr) (r *ipRange, err error) {
	defer func() { err = errors.Annotate(err, "invalid ip range: %w") }()

	if !start.Less(end) {
		return nil, fmt.Errorf("start is greater than or equal to end")
	}

	// Make sure that both are 16 bytes long to simplify handling in
	// methods.
	startData, endData := start.As16(), end.As16()

	startInt := (&big.Int{}).SetBytes(startData[:])
	endInt := (&big.Int{}).SetBytes(endData[:])
	diff := (&big.Int{}).Sub(endInt, startInt)

	if !diff.IsUint64() || diff.Uint64() > maxRangeLen {
		return nil, fmt.Errorf("range is too large")
	}

	return &ipRange{
		start: startInt,
		end:   endInt,
	}, nil
}

// contains returns true if r contains ip.
func (r *ipRange) contains(ip netip.Addr) (ok bool) {
	if r == nil {
		return false
	}

	ipData := ip.As16()

	return r.containsInt((&big.Int{}).SetBytes(ipData[:]))
}

// containsInt returns true if r contains ipInt.  For internal use only.
func (r *ipRange) containsInt(ipInt *big.Int) (ok bool) {
	return ipInt.Cmp(r.start) >= 0 && ipInt.Cmp(r.end) <= 0
}

// ipPredicate is a function that is called on every IP address in
// (*ipRange).find.  ip is given in the 16-byte form.
type ipPredicate func(ip netip.Addr) (ok bool)

// find finds the first IP address in r for which p returns true.  ip is in the
// 16-byte form.  It returns an empty [netip.Addr] if no addresses satisfy p.
func (r *ipRange) find(p ipPredicate) (ip netip.Addr) {
	if r == nil {
		return netip.Addr{}
	}

	_1 := big.NewInt(1)
	var ipData [16]byte
	for i := (&big.Int{}).Set(r.start); i.Cmp(r.end) <= 0; i.Add(i, _1) {
		i.FillBytes(ipData[:])
		ip = netip.AddrFrom16(ipData)
		if p(ip) {
			return ip
		}
	}

	return netip.Addr{}
}

// offset returns the offset of ip from the beginning of r.  It returns 0 and
// false if ip is not in r.
func (r *ipRange) offset(ip netip.Addr) (offset uint64, ok bool) {
	if r == nil {
		return 0, false
	}

	ipData := ip.As16()
	ipInt := (&big.Int{}).SetBytes(ipData[:])
	if !r.containsInt(ipInt) {
		return 0, false
	}

	offsetInt := (&big.Int{}).Sub(ipInt, r.start)

	// Assume that the range was checked against maxRangeLen during
	// construction.
	return offsetInt.Uint64(), true
}

// String implements the fmt.Stringer interface for *ipRange.
func (r *ipRange) String() (s string) {
	start, end := [16]byte{}, [16]byte{}

	r.start.FillBytes(start[:])
	r.end.FillBytes(end[:])

	return fmt.Sprintf("%s-%s", net.IP(start[:]), net.IP(end[:]))
}
