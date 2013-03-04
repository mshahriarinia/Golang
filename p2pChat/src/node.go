package main

/**
This project will implement a Peer-to-Peer command-line chat in Go language. 
@March 2013
@by Morteza Shahriari Nia   @mshahriarinia


Reading arbitrary strings from command-line was a bit trickey as I couldn't get a straight-forward example 
on telling how to do it. But after visiting tens of pages and blogs it was fixed.

Multi-threading communciation via channels is very useful but not in our context. We need Mutex 
for handling our clients which is not straightforward or natural in channels and message streams.  
*/

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strconv"
	//	"strings"
	"container/list"
	"sync"
	"time"
)

var (
	port            string
	serverIP               = "localhost" //TODO fix server ip
	SERVER_PORT     string = "5555"      //default port as the main p2p server
	stop                   = false
	mutexClientList sync.Mutex
)

func main() {
	//initialize values
	reader := bufio.NewReader(os.Stdin) //read line from standard input
	connList := list.New()              //list of p2p chat users.

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

	go acceptClients(port, connList)
	go chatSay(connList)
	runtime.Gosched() //let the new thread to start, otherwuse it will not execute.

	fmt.Println("\nStarting to read user inputs to chat.")

	//it's good to not include accepting new clients from main just in case the user
	//wants to quit by typing some keywords, the main thread is not stuck at
	// net.listen.accept forever
	for !stop {
		time.Sleep(1000 * time.Millisecond)
	} //keep main thread alive
}

//TODO maintain list of all nodes and send to everybody
//read access to channel list
//close the connection

func chatSay(connList *list.List) {
	reader := bufio.NewReader(os.Stdin) //get teh reader to read lines from standard input

	//conn, err := net.Dial("tcp", serverIP+":"+SERVER_PORT)

	for !stop { //keep reading inputs forever
		fmt.Print("Enter text to chat with the p2p network: ")
		str, _ := reader.ReadString('\n')

		mutexClientList.Lock()
		for e := connList.Front(); e != nil; e = e.Next() {
			conn := e.Value.(net.TCPConn)
			_, err := conn.Write([]byte(str)) //transmit string as byte array
			if err != nil {
				fmt.Println("Error send reply:", err.Error())
			}
		}
		mutexClientList.Unlock()
		fmt.Println("HOME" + ": " + str)
	}
}

//TODO at first get list of clients. be ready to get a new client any time
/**
Accept new clients. 
*/
func acceptClients(port string, connList *list.List) {
	fmt.Println("Listenning to port", port)

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error listenning to port ", port)
	}
	for !stop {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error in accepting connection.")
			continue
		}

		mutexClientList.Lock()
		connList.PushBack(conn)
		mutexClientList.Unlock()

		go handleClient(conn, connList)
		runtime.Gosched()
	}
}

/**
Receive message from client. listen and wait for content from client. the write to 
client will be performed when the current user enters an input
*/
func handleClient(conn net.Conn, connList *list.List) {
	fmt.Println("Handling Connection")

	buffer := make([]byte, 1024)
	for !stop {
		bytesRead, error := conn.Read(buffer)
		if error != nil {
			fmt.Println("Error in reading from connection")
		}
		input := string(buffer[0:bytesRead])
		fmt.Println(conn, input)
	}
}

/**
Generate a port number
*/
func generatePortNo() string {
	rand.Seed(time.Now().Unix())
	return strconv.Itoa(rand.Intn(5000) + 5000) //generate a valid port
}

/**
Determine the local IP addresses
*/
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
