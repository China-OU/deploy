package common

import (
	"net"
	"fmt"
)

func GetLocalIp() []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err.Error())
		return []string{}
	}

	var ip_arr []string
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip_arr = append(ip_arr, ipnet.IP.String())
			}
		}
	}
	return ip_arr
}
