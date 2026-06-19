package httpx

import (
	"fmt"
	"net/netip"
)

// maxSourceIPs limits the number of addresses after CIDR expansion to
// prevent memory exhaustion from large prefixes.
const maxSourceIPs = 1024

func expandLocalAddresses(input interface{}) ([]string, error) {
	if input == nil {
		return nil, nil
	}

	var raw []string
	switch v := input.(type) {
	case string:
		raw = []string{v}
	case []interface{}:
		raw = make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("source_ips element must be a string, got %T", item)
			}
			raw = append(raw, s)
		}
	default:
		return nil, fmt.Errorf("source_ips must be string or list, got %T", input)
	}

	var result []string
	for _, item := range raw {
		if containsSlash(item) {
			prefix, err := netip.ParsePrefix(item)
			if err != nil {
				return nil, fmt.Errorf("invalid CIDR prefix %q: %w", item, err)
			}
			addr := prefix.Addr()
			if !addr.Is4() && !addr.Is6() {
				return nil, fmt.Errorf("unsupported address family in %q", item)
			}

			// Skip network address (first) and, for IPv4, broadcast address (last)
			skipFirst := true
			for prefix.Contains(addr) {
				if skipFirst {
					skipFirst = false
					addr = addr.Next()
					continue
				}
				// Check if this is the last address in the prefix
				next := addr.Next()
				lastAddr := !next.IsValid() || !prefix.Contains(next)
				if lastAddr && addr.Is4() && prefix.Bits() < 31 {
					break // skip broadcast for IPv4 prefixes /30 and larger
				}
				result = append(result, addr.String())
				if len(result) > maxSourceIPs {
					return nil, fmt.Errorf("too many source_ips after CIDR expansion (%d > %d)", len(result), maxSourceIPs)
				}
				if lastAddr {
					break
				}
				addr = next
			}
		} else {
			addr, err := netip.ParseAddr(item)
			if err != nil {
				return nil, fmt.Errorf("invalid IP address %q: %w", item, err)
			}
			result = append(result, addr.String())
		}
	}

	return result, nil
}

func containsSlash(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return true
		}
	}
	return false
}
