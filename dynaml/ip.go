package dynaml

import (
	"fmt"
	"net"
	"slices"
	"strings"

	"github.com/mandelsoft/spiff/yaml"
)

func func_ip(op func(ip net.IP, cidr *net.IPNet) interface{}, arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		info.Issue = yaml.NewIssue("only one argument expected for CIDR function")
		return nil, info, false
	}

	str, ok := arguments[0].(string)
	if !ok {
		info.Issue = yaml.NewIssue("CIDR argument required")
		return nil, info, false
	}

	ip, cidr, err := net.ParseCIDR(str)

	if err != nil {
		info.Issue = yaml.NewIssue("CIDR argument required")
		return nil, info, false
	}

	return op(ip, cidr), info, true
}

func func_containsIP(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 2 {
		info.Issue = yaml.NewIssue("contains_ip requires CIDR and IP argument")
		return nil, info, false
	}

	str, ok := arguments[0].(string)
	if !ok {
		info.Issue = yaml.NewIssue("CIDR required as first argument")
		return nil, info, false
	}

	_, cidr, err := net.ParseCIDR(str)

	if err != nil {
		info.Issue = yaml.NewIssue("CIDR argument required: %s", err)
		return nil, info, false
	}

	str, ok = arguments[1].(string)
	if !ok {
		info.Issue = yaml.NewIssue("IP required as second argument")
		return nil, info, false
	}

	ip := net.ParseIP(str)
	if ip == nil {
		info.Issue = yaml.NewIssue("IP argument required: %s", str)
		return nil, info, false
	}
	return cidr.Contains(ip), info, true
}

func func_minIP(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	return func_ip(func(ip net.IP, cidr *net.IPNet) interface{} {
		return ip.Mask(cidr.Mask).String()
	}, arguments, binding)
}

func func_maxIP(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	return func_ip(func(ip net.IP, cidr *net.IPNet) interface{} {
		return MaxIP(cidr).String()
	}, arguments, binding)
}

func func_numIP(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	return func_ip(func(ip net.IP, cidr *net.IPNet) interface{} {
		ones, _ := cidr.Mask.Size()
		return int64(1 << (32 - uint32(ones)))
	}, arguments, binding)
}

func SubIP(ip net.IP, mask net.IPMask) net.IP {
	m := ip.Mask(mask)
	out := make(net.IP, len(ip))
	for i, v := range ip {
		j := len(ip) - i
		if j > len(m) {
			out[i] = v
		} else {
			out[i] = v &^ m[len(m)-j]
		}
	}
	return out
}

func MaxIP(cidr *net.IPNet) net.IP {
	ip := cidr.IP.Mask(cidr.Mask)
	mask := cidr.Mask
	out := make(net.IP, len(ip))
	for i, v := range ip {
		j := len(ip) - i
		if j > len(mask) {
			out[i] = v
		} else {
			out[i] = v | ^mask[len(mask)-j]
		}
	}
	return out
}

func DiffIP(a, b net.IP) int64 {
	var d int64

	for i, _ := range a {
		db := int64(a[i]) - int64(b[i])
		d = d*256 + db
	}
	return d
}

////////////////////////////////////////////////////////////////////////

func ifceselector(name string, arguments []interface{}) (int, []net.Interface, error) {
	typ := 0
	var filter []string

	for _, arg := range arguments {
		s, ok := arg.(string)
		if !ok {
			return 0, nil, fmt.Errorf("argument to %s must be a string", name)
		}
		for _, v := range strings.Split(s, ",") {
			v = strings.TrimSpace(v)
			switch v {
			case "v4":
				typ |= 1
			case "v6":
				typ |= 2
			case "loopback":
				typ |= 4
			default:
				filter = append(filter, v)
			}
		}
	}
	if typ == 0 {
		typ = 5
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return 0, nil, fmt.Errorf("cannot get network interfaces: %s", err.Error())
	}

	var result []net.Interface
	for _, iface := range interfaces {
		// Filter out interfaces that aren't useful
		// Skip if the interface is down or doesn't have a name
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		if filter != nil && !slices.Contains(filter, iface.Name) {
			continue
		}
		result = append(result, iface)
	}
	return typ, result, nil
}

func func_localips(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if !binding.GetState().OSAccessAllowed() {
		return info.DenyOSOperation("local_ips")
	}

	typ, interfaces, err := ifceselector("localIPs", arguments)
	if err != nil {
		return info.Error("%s", err.Error())
	}

	var result []yaml.Node

	for _, iface := range interfaces {
		// Get addresses associated with this specific interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			// Use type assertion to get the IP
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil {
				continue
			}

			if ip.IsLoopback() && (typ&4 == 0) {
				continue
			}

			if ip.To4() != nil {
				if typ&1 != 0 {
					result = append(result, yaml.NewNode(ip.String(), binding.SourceName()))
				}
			} else {
				if typ&2 != 0 {
					result = append(result, yaml.NewNode(ip.String(), binding.SourceName()))
				}
			}
		}
	}
	return result, info, true
}

func func_localinterfaces(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if !binding.GetState().OSAccessAllowed() {
		return info.DenyOSOperation("local_interfaces")
	}

	typ, interfaces, err := ifceselector("localInterfaces", arguments)
	if err != nil {
		return info.Error("%s", err.Error())
	}

	result := map[string]yaml.Node{}

	for _, iface := range interfaces {
		r := map[string]yaml.Node{}

		// Get addresses associated with this specific interface
		a := []yaml.Node{}
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				if ip == nil {
					continue
				}

				if ip.To4() != nil {
					if typ&1 != 0 {
						a = append(a, yaml.NewNode(ip.String(), binding.SourceName()))
					}
				} else {
					if typ&2 != 0 {
						a = append(a, yaml.NewNode(ip.String(), binding.SourceName()))
					}
				}
			}
		}
		r["addrs"] = yaml.NewNode(a, binding.SourceName())
		r["name"] = yaml.NewNode(iface.Name, binding.SourceName())
		r["mtu"] = yaml.NewNode(iface.MTU, binding.SourceName())
		r["index"] = yaml.NewNode(iface.Index, binding.SourceName())
		r["mac"] = yaml.NewNode(iface.HardwareAddr.String(), binding.SourceName())

		result[iface.Name] = yaml.NewNode(r, binding.SourceName())
	}
	return result, info, true
}
