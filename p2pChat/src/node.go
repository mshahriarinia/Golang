package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

var (
	port       int  
	serverIP       = "localhost"     //TODO fix server ip
	SERVER_PORT int = 5555  //default port as the main p2p server
)

func main() {
	//initialize values
	reader := bufio.NewReader(os.Stdin) //read line from standard input

	fmt.Println("               Welcome to Peer-to-Peer (P2P) Command-Line Chat in Go language.")
	fmt.Print("Run this node as main server? (y/n) ")

	str, err := reader.ReadString('\n') //ignore the error by sending it to nil
	if err != nil {
		fmt.Println("Can not read from command line.")
		os.Exit(1)
	}

	if []byte(str)[0] == 'y' {
		fmt.Println("Starting the node as the main p2p server.")
		port = SERVER_PORT
	} else if []byte(str)[0] == 'n' {
		fmt.Println("Starting the node as a normal p2p node.")
		port = generatePortNo()
	} else {
		fmt.Println("Wrong argument type.")
		os.Exit(1)
	}

	fmt.Println("Server Socket: " + serverIP + ":" + strconv.Itoa(SERVER_PORT))

	localIp := getLocalIP()
	fmt.Println("Local Socket: " + localIp[0] + ":" + strconv.Itoa(port))

	go listenToPort(port)
	fmt.Println("After Go.")
	

}

func listenToPort(port int) {
	fmt.Println("Listenning to port", port)
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Error listenning to port ", port)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error in accepting connection.")
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	fmt.Println("Handling Connection")
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
	fmt.Println("Local Hostname: " + name)

	addrs, err := net.LookupHost(name)
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return []string{}
	}
	fmt.Println("Local IP Addresses: ", addrs)

	//for _, a := range addrs {    //print addresses
	//fmt.Println(a)
	//}
	return addrs
}
