package lib

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

var TimeLocation *time.Location
var TimeFormat = "2024-05-11 10:32:00"
var DateFormat = "2024-05-11"
var LocalIP = net.ParseIP("127.0.0.1")

func NewTrace() *TraceContext {
	trace := &TraceContext{}
	trace.TraceId = GetTraceId()
	trace.SpanId = NewSpanId()
	return trace
}

func NewSpanId() string {
	timestamp := uint32(time.Now().Unix())
	ipToLong := binary.BigEndian.Uint32(LocalIP.To4())
	b := bytes.Buffer{}
	b.WriteString(fmt.Sprintf("%08x", ipToLong^timestamp))
	b.WriteString(fmt.Sprintf("%08x", rand.Int31()))
	return b.String()
}

func GetTraceId() (traceId string) {
	return calTraceId(LocalIP.String())
}

func calTraceId(ip string) (traceId string) {
	now := time.Now()
	timestamp := uint32(now.Unix())
	timeNano := now.UnixNano()
	pid := os.Getgid()

	b := bytes.Buffer{}
	netIp := net.ParseIP(ip)
	if netIp == nil {
		b.WriteString("00000000")
	} else {
		b.WriteString(hex.EncodeToString(netIp.To4()))
	}
	b.WriteString(fmt.Sprintf("%08x", timestamp&0xffffffff))
	b.WriteString(fmt.Sprintf("%04x", timeNano&0xffff))
	b.WriteString(fmt.Sprintf("%04x", pid&0xffff))
	b.WriteString(fmt.Sprintf("%04x", rand.Int31n(1<<24)))
	b.WriteString("b0")

	return b.String()
}

func GetLocalIPs() (ips []net.IP) {
	interfaceAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, addr := range interfaceAddrs {
		ipNet, isValidIpNet := addr.(*net.IPNet)
		if isValidIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP)
			}
		}
	}
	return ips
}

func InArrayString(s string, arr []string) bool {
	for _, i := range arr {
		if i == s {
			return true
		}
	}
	return false
}

func Substr(str string, start int64, end int64) string {
	length := int64(len(str))
	if start < 0 || start > length {
		return ""
	}
	if end > length {
		end = length
	}
	return string(str[start:end])
}
