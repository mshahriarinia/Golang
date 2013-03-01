package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

var (
	id             = "Server"
	port       int = 5555
	serverIP       = "localhost"
	serverPort     = 5555
)

func main() {

	//read line from standard input
	reader := bufio.NewReader(os.Stdin)
	str, err := reader.ReadString('\n')
	fmt.Println(str, err)

	fmt.Println("Id: ", id, "IP: ")
	if id != "Server" {
		port = generatePortNo()
	}
	fmt.Println(port)
	fmt.Println("Node Starting w/ ip")
	localIp := getLocalIP()
	fmt.Println("Local Socket: ", localIp, ":", port)
	fmt.Println("Server Socket: ", serverIP, ":", serverPort)
}

func generatePortNo() int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(5000) + 5000 //generate a valid port
}

func getLocalIP() []string {
	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return []string{}
	}
	fmt.Println("Hostname: ", name)
	addrs, err := net.LookupHost(name)
	fmt.Println(addrs)
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return []string{}
	}
	for _, a := range addrs {
		fmt.Println(a)
	}
	return addrs
}
