package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/kardianos/service"
)

var allMacAddrs []net.HardwareAddr

var isRuning bool = false

func getMacAddrs() (macAddrs []net.HardwareAddr) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v", err)
		return macAddrs
	}
	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr
		if len(macAddr) == 0 {
			continue
		}
		macAddrs = append(macAddrs, macAddr)
	}
	return macAddrs
}
func findAnyMac(macBytes []byte) bool {
	for _, item := range allMacAddrs {
		if bytes.Equal(item, macBytes) {
			return true
		}
	}
	return false
}

func udpServer(address string) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		logger.Errorf("ResolveUDPAddr failed,err:" + err.Error())
		return
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logger.Errorf("ListenUDP failed,err:" + err.Error())
		return
	}
	defer conn.Close()
	if err != nil {
		logger.Errorf("read from connect  failed,err:" + err.Error())
		return
	}
	logger.Infof("started listen upd %v", address)
	data := make([]byte, 128)
	for {
		if !isRuning {
			break
		}
		n, _, err := conn.ReadFromUDP(data)
		if err != nil {
			logger.Errorf("failed read udp msg, error:%v ", err.Error())
			return
		}
		if n != 102 {
			logger.Errorf("recv udp msg was %v bytes,(expected 102 bytes sent)", n)
			return
		}
		macBytes := data[6:12]
		if !findAnyMac(macBytes) {
			logger.Errorf("recv magic msg but mac address not match")
			return
		}
		//关闭电脑
		if data[0] == 0x00 {
			os.Getenv("windir")
			cmd := exec.Command(os.Getenv("windir")+"\\system32\\rundll32.exe", "powrprof.dll,SetSuspendState")
			d, err := cmd.CombinedOutput()
			if err != nil {
				logger.Errorf("Hibernate Error:", err)
				return
			}
			logger.Infof("Hibernate result:" + string(d))
		}
	}
}

var logger service.Logger

type program struct{}

func (p *program) Start(s service.Service) error {
	isRuning = true
	go p.run()
	return nil
}
func (p *program) run() {
	allMacAddrs = getMacAddrs()
	udpServer("0.0.0.0:9")
}
func (p *program) Stop(s service.Service) error {
	isRuning = false
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "remote-shutdown-service",
		DisplayName: "remote-shutdown-service",
		Description: "远程关机服务",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) < 2 {
		err = s.Run()
		if err != nil {
			logger.Error(err)
		}
		return
	}
	cmd := os.Args[1]
	if cmd == "install" {
		err = s.Install()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("安装成功")
	}
	if cmd == "uninstall" {
		err = s.Uninstall()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("卸载成功")
	}
}
