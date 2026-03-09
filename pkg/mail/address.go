package mail

import "strings"

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
