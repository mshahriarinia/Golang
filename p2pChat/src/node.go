package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strconv"
//	"strings"
	"time"
	"container/list"
)

var (
	port        string
	serverIP           = "localhost" //TODO fix server ip
	SERVER_PORT string = "5555"      //default port as the main p2p server
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

	fmt.Println("Server Socket: " + serverIP + ":" + SERVER_PORT)

	localIp := getLocalIP()
	fmt.Println("Local Socket: " + localIp[0] + ":" + port)

	go chatListen(port)
	go chatSay()
	runtime.Gosched() //let the new thread to start, otherwuse it will not execute.

	fmt.Println("\nStarting to read user inputs to chat.")

	for {
		time.Sleep(1000 * time.Millisecond)
	} //keep main thread alive
}


//TODO maintain list of all nodes and send to everybody
func chatSay() {
	reader := bufio.NewReader(os.Stdin) //read line from standard input

	conn, err := net.Dial("tcp", serverIP+":"+SERVER_PORT)
	for { //keep reading inputs forever
		fmt.Print("Enter text to chat with the p2p network: ")
		str, _ := reader.ReadString('\n')
		_, err = conn.Write([]byte(str)) //transmit string as byte array
		if err != nil {
			fmt.Println("Error send reply:", err.Error())
		} 
		fmt.Println("HOME" + ": " + str)
	}
}

//TODO at first get list of clients. be ready to get a new client information any time
//we need special message format for it
func chatListen(port string) {
	fmt.Println("Listenning to port", port)
	lst := list.New() //list of p2p chat users.
	
	ln, err := net.Listen("tcp", ":"+port)
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
		runtime.Gosched()
	}
}


func handleConnection(conn net.Conn) {
	fmt.Println("Handling Connection")
}

func generatePortNo() string {
	rand.Seed(time.Now().Unix())
	return strconv.Itoa(rand.Intn(5000) + 5000) //generate a valid port
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
