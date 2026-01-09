package utils

import "strconv"

func TransportString(method, host string, port *uint16, mxLookup bool) string {
	str := method + ":"
	if !mxLookup {
		str += "[" + host + "]"
	} else {
		str += host
	}
	if port != nil {
		str += ":" + strconv.Itoa(int(*port))
	}
	return str
}
