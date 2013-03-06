package main

/**
This project will implement a Peer-to-Peer command-line chat in Go language. 
@March 2013
@by Morteza Shahriari Nia   @mshahriarinia


//in one control message only send the actual port

Reading arbitrary strings from command-line was a bit trickey as I couldn't get a straight-forward example 
on telling how to do it. But after visiting tens of pages and blogs it was fixed. Buffers, buffered reader, 
streams, ... The diffference between when you learn womething and when you actually do it.

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
	"strings"
	//	"strings"
	"container/list"
	"sync"
	"time"
)

var (
	port                     string
	SERVER_IP                       = "127.0.0.1" //TODO fix server ip
	SERVER_PORT              string = "5555"      //default port as the main p2p server
	stop                            = false
	mutexPeerList          sync.Mutex
	CONTROL_MESSAGE_PREAMBLE = "\u001B" + ":!q" //char code used in VIM to exit the program
)

type Peer struct {
		conn net.Conn
		port string
		ip string
	}

func main() {
	//initialize values
	reader := bufio.NewReader(os.Stdin) //read line from standard input
	peerList := list.New()            //list of p2p chat users.

	fmt.Println("\n\n               Welcome to Peer-to-Peer (P2P) Command-Line Chat in Go language.\n\n")
	fmt.Print("Run this node as main server? (y/n) ")

	str, err := reader.ReadString('\n') //ignore the error by sending it to nil
	if err != nil {
		fmt.Println("Can not read from command line.")
		os.Exit(1)
	}

	if []byte(str)[0] == 'y' {
		fmt.Println("Node is the main p2p server.")
		port = SERVER_PORT
	} else if []byte(str)[0] == 'n' {
		fmt.Println("Node is a normal p2p node.")
		port = generatePortNo()
		connectToIpPort(SERVER_IP+":"+SERVER_PORT, peerList)
	} else {
		fmt.Println("Wrong argument type.")
		os.Exit(1)
	}

	fmt.Println("Server Socket: " + SERVER_IP + ":" + SERVER_PORT)
	localIp := getLocalIP()
	fmt.Println(" Local Socket: " + localIp[0] + ":" + port)
	fmt.Println("---------------------------------------------------------")

	go acceptPeers(port, peerList)
	go chatSay(peerList)
	
//	if []byte(str)[0] == 'n' {
//		connectToPeer(SERVER_IP+":"+SERVER_PORT, peerList)
//	}
	runtime.Gosched() //let the new thread to start, otherwuse it will not execute.

	//it's good to not include accepting new clients from main just in case the user
	//wants to quit by typing some keywords, the main thread is not stuck at
	// net.listen.accept forever
	for !stop {
		time.Sleep(1000 * time.Millisecond)
	} //keep main thread alive
}

func connectToPeers(peer Peer, controlMessage string, peerList *list.List) {
	strArr := strings.Split(controlMessage, " ")
	for i, ipport := range strArr {
		if i == 0 {
			//skip preamble
		} else if i ==1 { //set actual port for the peer sending this message
			peer.port = ipport
		}else if !isSelf(ipport) { //skip preamble
			connectToIpPort(ipport, peerList)
		}
	}
}

/**
ask for a connection from a node via ipport
*/
func connectToIpPort(ipport string, peerList *list.List) {
	if strings.Contains(ipport, "nil"){
		return
	}
	  
	mutexPeerList.Lock()
	conn, err := net.Dial("tcp", ipport)	 
	if err != nil {
		fmt.Println("Error connecting to:", ipport, err.Error())
		mutexPeerList.Unlock()
		return
		
	}
	peer := &Peer{conn, "nilport", getIP(conn)}
	
	peerList.PushBack(peer)
	mutexPeerList.Unlock()
	
	go handlePeer(peer, peerList)
	runtime.Gosched()
}


//TODO maintain list of all nodes and send to everybody
//read access to channel list
//close the connection

func chatSay(peerList *list.List) {
	reader := bufio.NewReader(os.Stdin) //get teh reader to read lines from standard input

	//conn, err := net.Dial("tcp", serverIP+":"+SERVER_PORT)

	for !stop { //keep reading inputs forever
		fmt.Print("user@Home[\\ ")
		str, _ := reader.ReadString('\n')

		mutexPeerList.Lock()
		for e := peerList.Front(); e != nil; e = e.Next() {
			conn := e.Value.(*Peer).conn
			_, err := conn.Write([]byte(str)) //transmit string as byte array
			if err != nil {
				fmt.Println("Error sending reply:", err.Error())
			}
		}
		mutexPeerList.Unlock()
	}
}

//TODO close connections
//TODO forward new ip:port to other nodes

//TODO at first get list of clients. be ready to get a new client any time
/**
Accept new clients. 
*/
func acceptPeers(port string, peerList *list.List) {
	//fmt.Println("Listenning to port", port)
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error listenning to port ", port)
		stop = true
	}
	for !stop {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error in accepting connection.")
			stop = true
			continue
		}
		
		mutexPeerList.Lock()
		peer := &Peer{conn, "nilport", getIP(conn)}  
		peerList.PushBack(peer)
		mutexPeerList.Unlock()

		go handlePeer(peer, peerList)
		runtime.Gosched()
	}
}

/**
Receive message from client. 
Listen and wait for content from client. the write to 
client will be performed when the current user enters an input
*/
func handlePeer(peer *Peer, peerList *list.List) {
	stopConn := false
	fmt.Println("New node: ", peer.conn.RemoteAddr())
		
		
	//send current peer list
	str := peerListToStr(peerList) 
	_, err := peer.conn.Write([]byte(CONTROL_MESSAGE_PREAMBLE + " " + port + " " + str)) //transmit string as byte array
	if err != nil {
		fmt.Println("Error sending reply:", err.Error())
	}	
	
	//Listen for the peer messages
	buffer := make([]byte, 1024)

	for !stopConn {
		bytesRead, err := peer.conn.Read(buffer)
		if err != nil { //stop for loop, remove peer from list
			
			stopConn = true
			fmt.Println("Error in reading from connection", peer.conn.RemoteAddr())
			mutexPeerList.Lock()
			el := getListElement(*peer, peerList)
			if el != nil {
				peerList.Remove(el)
			}
			mutexPeerList.Unlock()
			
		} else {
			messageStr := string(buffer[0:bytesRead])
			fmt.Println(peer.conn.RemoteAddr(), " says: ", messageStr)

			if strings.Contains(messageStr, CONTROL_MESSAGE_PREAMBLE) {
				//pass peer itself to set actual port
				sArr := strings.Split(messageStr, " ")
				fmt.Println("port isSSSSSS: ", sArr[1])
				
				
				el := getListElement(*peer, peerList)
				p := el.Value.(*Peer)
				p.port = sArr[1]
				//peer.port = sArr[1]  
				fmt.Println("setted port to", p.port)
				setPort(*peer, peerList, sArr[1])
				
				connectToPeers(*peer, messageStr, peerList) 
				printlist(peerList)
			}
		}
	}
	fmt.Println("Closing ", peer.conn.RemoteAddr())
	peer.conn.Close()
}


func setPort(peer Peer, l *list.List, port string) *list.Element {
	for e := l.Front(); e != nil; e = e.Next() {
		temp := e.Value.(*Peer)
		
		if peer.conn.RemoteAddr() == temp.conn.RemoteAddr() {
			fmt.Println("found connection.")
			temp.port = port
			return e
		}
	}
	return nil
}


func getListElement(peer Peer, l *list.List) *list.Element {
	for e := l.Front(); e != nil; e = e.Next() {
		temp := e.Value.(*Peer)
		
		if peer.conn.RemoteAddr() == temp.conn.RemoteAddr() {
			fmt.Println("found connection.")
			return e
		}
	}
	return nil
}


/**
Get a string of the peer list as ip:port
*/
func peerListToStr(l *list.List) string {
	if l == nil {
		return ""
	}
	s := ""
	mutexPeerList.Lock()
	for e := l.Front(); e != nil; e = e.Next() {
		peer := e.Value.(*Peer)
		s = s + peer.ip + ":" + peer.port + " "
	}
	//s = s + getLocalIP()[0] + ":" + port
	mutexPeerList.Unlock()
	return strings.Trim(s, " ")
}

func printlist(l *list.List) {
	fmt.Print("\nConnection List: [")
	fmt.Print(peerListToStr(l))
	fmt.Println("]")
}

func (p *Peer) ipport() string{
	return p.ip + ":" + p.port
}

/**
Checks to see if the ipport combination is the current node itself. 
*/
func isSelf(ipport string) bool {
	ipArr := getLocalIP()

	for _, ip := range ipArr {
		if ipport == ip+":"+port {
			return true
		}
	}
	if ipport == "127.0.0.1"+":"+port || ipport == "localhost"+":"+port {
		return true
	}
	return false
}

/**
Generate a port number
*/
func generatePortNo() string {
	rand.Seed(time.Now().Unix())
	return strconv.Itoa(rand.Intn(5000) + 5000) //generate a valid port
}

func getIP(conn net.Conn) string{
	s := conn.RemoteAddr().String()
	s = strings.Split(s, ":")[0]
	s = strings.Trim(s, ":")
	return s
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
	//fmt.Print("Local Hostname: " + name)

	addrs, err := net.LookupHost(name)
	if err != nil {
		fmt.Printf("Oops: %v\n", err)
		return []string{}
	}
	//	fmt.Print("\t\tLocal IP Addresses: ", addrs)

	return addrs
}
