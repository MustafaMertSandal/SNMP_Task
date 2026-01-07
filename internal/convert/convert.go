package convert

import (
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"

	"github.com/gosnmp/gosnmp"
)

func PDUToString(pdu gosnmp.SnmpPDU) string {
	switch v := pdu.Value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}

func PDUToBigInt(pdu gosnmp.SnmpPDU) *big.Int {
	return gosnmp.ToBigInt(pdu.Value)
}

func PDUToUint64(pdu gosnmp.SnmpPDU) uint64 {
	bi := PDUToBigInt(pdu)
	if bi.Sign() < 0 {
		return 0
	}
	return bi.Uint64()
}

func PDUToInt(pdu gosnmp.SnmpPDU) int {
	return int(PDUToBigInt(pdu).Int64())
}

func PDUToIPv4(pdu gosnmp.SnmpPDU) string {
	switch v := pdu.Value.(type) {
	case string:
		return v
	case []byte:
		if len(v) == 4 {
			return net.IP(v).String()
		}
		return strings.TrimSpace(string(v))
	default:
		return fmt.Sprint(v)
	}
}

func ParseLastIntIndex(oid string) (int, error) {
	parts := strings.Split(strings.Trim(oid, "."), ".")
	if len(parts) == 0 {
		return 0, fmt.Errorf("empty oid")
	}
	return strconv.Atoi(parts[len(parts)-1])
}

func ParseLastIPv4FromOID(oid string) (string, error) {
	parts := strings.Split(strings.Trim(oid, "."), ".")
	if len(parts) < 4 {
		return "", fmt.Errorf("oid too short for ipv4: %s", oid)
	}
	n := len(parts)
	a, _ := strconv.Atoi(parts[n-4])
	b, _ := strconv.Atoi(parts[n-3])
	c, _ := strconv.Atoi(parts[n-2])
	d, _ := strconv.Atoi(parts[n-1])
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d), nil
}

func ParseLastNInts(oid string, n int) ([]int, error) {
	parts := strings.Split(strings.Trim(oid, "."), ".")
	if len(parts) < n {
		return nil, fmt.Errorf("oid too short: %s", oid)
	}
	out := make([]int, 0, n)
	for i := len(parts) - n; i < len(parts); i++ {
		v, err := strconv.Atoi(parts[i])
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func OperStatusText(v int) string {
	switch v {
	case 1:
		return "up"
	case 2:
		return "down"
	case 3:
		return "testing"
	case 4:
		return "unknown"
	case 5:
		return "dormant"
	case 6:
		return "notPresent"
	case 7:
		return "lowerLayerDown"
	default:
		return "n/a"
	}
}

func RouteTypeText(v int) string {
	switch v {
	case 1:
		return "other"
	case 2:
		return "invalid"
	case 3:
		return "direct"
	case 4:
		return "indirect"
	default:
		return "n/a"
	}
}
