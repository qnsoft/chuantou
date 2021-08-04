package test

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

type CnIp struct {
	value int
	max   int
}

func (p *CnIp) Contains(ip int) bool {
	return ip >= p.value && ip <= p.value+p.max
}

func (p *CnIp) Format() string {
	return fmt.Sprintf("%s~%s", p.print(p.value), p.print(p.value+p.max))
}

func (p *CnIp) print(ip int) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ip>>24, (ip&0xFFFFFF)>>16, (ip&0xFFFF)>>8, ip&0xFF)
}

func IpToNumber(ip string) (result int, err error) {
	slices := strings.Split(ip, ".")
	length := len(slices)
	if length != 4 {
		return -1, errors.New("illegal ip string length")
	}

	for index, slice := range slices {
		i, _ := strconv.Atoi(slice)
		if i < 0 || i >= 256 {
			return -1, errors.New("illegal ip strings")
		}
		result += i * int(math.Pow(256, float64(length-index-1)))
	}
	return result, nil
}

var CnIpPool = make([]CnIp, 0)

func init() {
	fi, err := os.Open("../cn_ips.txt")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fi.Close()
	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		lines := strings.Split(string(a), ",")
		max, _ := strconv.Atoi(strings.TrimSpace(lines[1]))

		ip, _ := IpToNumber(lines[0])
		local := CnIp{ip, max}
		CnIpPool = append(CnIpPool, local)
	}
	fmt.Println("Init CnIp success. size", len(CnIpPool))
}

func IsCnIp(ip string) bool {
	if ipNumber, err := IpToNumber(ip); err == nil {
		for _, cnIp := range CnIpPool {
			if cnIp.Contains(ipNumber) {
				//fmt.Println(cnIp.Format())
				return true
			}
		}
	}
	return false
}

func TestIpToNumber(t *testing.T) {
	fmt.Println(IpToNumber("39.184.149.118"))
	fmt.Println(IpToNumber("39.128.0.0"))
}

func TestContainsIp(t *testing.T) {
	fmt.Println("--国内IP判断--")
	var ips = []string{"39.184.149.118", "121.17.142.60", "45.141.87.9"}

	for _, ip := range ips {
		fmt.Printf("%s %t\n", ip, IsCnIp(ip))
	}
}

func TestContainsIp2(t *testing.T) {
	var ip, _ = IpToNumber("39.184.149.118")
	var value, _ = IpToNumber("39.128.0.0")
	var cnIp = CnIp{value, 4194304}

	fmt.Println(cnIp.Contains(ip))
}
