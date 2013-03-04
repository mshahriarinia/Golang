package main

import (
    "fmt";
    "net";
    "log";
    "os";
    "container/list";
    "strings";
    "bytes";
    "flag";
)

// flag for debuging info. or a simple log
var debug = flag.Bool("d", false, "set the debug modus( print informations )")

type ClientChat struct {
    Name string;        // name of user
    IN chan string;     // input channel for to send to user
    OUT chan string;    // input channel from user to all
    Con *net.Conn;      // connection of client
    Quit chan bool;     // quit channel for all goroutines
    ListChain *list.List;    // reference to list
}

// read from connection and return true if ok
func (c *ClientChat) Read(buf []byte) bool{
    nr, err := c.Con.Read(buf);
    if err!=nil {
        c.Close();
        return false;
    }
    Log("Read():  ", nr, " bytes");
    return true;
}

// close the connection and send quit to sender
func (c *ClientChat) Close() {
    c.Quit<-true;
    c.Con.Close();
    c.deleteFromList();
}

// compare two clients: name and network connection
func (c *ClientChat) Equal(cl *ClientChat) bool {
    if bytes.Equal(strings.Bytes(c.Name), strings.Bytes(cl.Name)) {
        if c.Con == cl.Con {
            return true;
        }
    }
    return false;
}

// delete the client from list
func (c *ClientChat) deleteFromList() {
    for e := c.ListChain.Front(); e != nil; e = e.Next() {
        client := e.Value.(ClientChat);
        if c.Equal(&client) {
            Log("deleteFromList(): ", c.Name);
            c.ListChain.Remove(e);
        }
    }
}

// func Log(v ...): loging. give log information if debug is true
func Log(v ...interface{}) {
    if *debug == true {
        ret := fmt.Sprint(v);
        log.Stdoutf("SERVER: %s", ret);
    }
}

// func test(): testing for error
func test(err os.Error, mesg string) {
    if err!=nil {
        log.Stderr("SERVER: ERROR: ", mesg);
         os.Exit(-1);
    } else{
        Log("Ok: ", mesg)
        }
}

// handlingINOUT(): handle inputs from client, and send it to all other client via channels.
func handlingINOUT(IN <-chan string, lst *list.List) {
    for {
        Log("handlingINOUT(): wait for input");
        input := <-IN;  // input, get from client
        // send to all client back
        Log("handlingINOUT(): handling input: ", input);
        for value := range lst.Iter() {
            client := value.(ClientChat);
            Log("handlingINOUT(): send to client: ", client.Name);
            client.IN<- input;
        }  
    }
}



// clientreceiver wait for an input from network, after geting data it send to
// handlingINOUT via a channel.
func clientreceiver(client *ClientChat) {
    buf := make([]byte, 2048);

    Log("clientreceiver(): start for: ", client.Name);
    for client.Read(buf) {
        
        if bytes.Equal(buf, strings.Bytes("/quit")) {
            client.Close();
            break;
        }
        Log("clientreceiver(): received from ",client.Name, " (", string(buf), ")");
        send := client.Name+"> "+string(buf);
        client.OUT<- send;
        for i:=0; i<2048;i++ {
            buf[i]=0x00;
        }
    }    

    client.OUT <- client.Name+" has left chat";
    Log("clientreceiver(): stop for: ", client.Name);
}

// clientsender(): get the data from handlingINOUT via channel (or quit signal from
// clientreceiver) and send it via network
func clientsender(client *ClientChat) {
    Log("clientsender(): start for: ", client.Name);
    for {
        Log("clientsender(): wait for input to send");
        select {
            case buf := <- client.IN:
                Log("clientsender(): send to \"", client.Name, "\": ", string(buf));
                client.Con.Write(strings.Bytes(buf));
            case <-client.Quit:
                Log("clientsender(): client want to quit");
                client.Con.Close();
                break;
        }
    }
    Log("clientsender(): stop for: ", client.Name);
}

// clientHandling(): get the username and create the clientsturct
// start the clientsender/receiver, add client to list.
func clientHandling(con *net.Conn, ch chan string, lst *list.List) {
    buf := make([]byte, 1024);
    con.Read(buf);
    name := string(buf);
    newclient := &ClientChat{name, make(chan string), ch, con, make(chan bool), lst};

    Log("clientHandling(): for ", name);
    go clientsender(newclient);
    go clientreceiver(newclient);
    lst.PushBack(*newclient);
    ch<- name+" has joinet the chat";
}

func main() {
    flag.Parse();
    Log("main(): start");

    // create the list of clients
    clientlist := list.New();
    in := make(chan string);
    Log("main(): start handlingINOUT()");
    go handlingINOUT(in, clientlist);
    
    // create the connection
    netlisten, err := net.Listen("tcp", "127.0.0.1:9988");
    test(err, "main Listen");
    defer netlisten.Close();

    for {
        // wait for clients
        Log("main(): wait for client ...");
        conn, err := netlisten.Accept();
        test(err, "main: Accept for client");
        go clientHandling(&conn, in, clientlist);
    }
}