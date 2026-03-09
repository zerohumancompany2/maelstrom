package mail

import (
	"fmt"
	"strings"
)

type AddressType int

const (
	AddressTypeAgent AddressType = iota
	AddressTypeTopic
	AddressTypeSys
)

func IsValidAgentAddress(addr string) bool {
	return isValidAddressWithPrefix(addr, "agent:")
}

func IsValidTopicAddress(addr string) bool {
	return isValidAddressWithPrefix(addr, "topic:")
}

func IsValidSysAddress(addr string) bool {
	return isValidAddressWithPrefix(addr, "sys:")
}

func isValidAddressWithPrefix(addr, prefix string) bool {
	return strings.HasPrefix(addr, prefix) && len(addr) > len(prefix)
}

func ParseAddress(addr string) (AddressType, string, error) {
	switch {
	case strings.HasPrefix(addr, "agent:"):
		return AddressTypeAgent, strings.TrimPrefix(addr, "agent:"), nil
	case strings.HasPrefix(addr, "topic:"):
		return AddressTypeTopic, strings.TrimPrefix(addr, "topic:"), nil
	case strings.HasPrefix(addr, "sys:"):
		return AddressTypeSys, strings.TrimPrefix(addr, "sys:"), nil
	default:
		return 0, "", fmt.Errorf("invalid address format: %s", addr)
	}
}
