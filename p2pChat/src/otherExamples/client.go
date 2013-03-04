package main

import (
    "fmt";
    "net";
    "log";
    "os";
    "bytes";
    "bufio";
    "strings";
    "time";
    "flag";
)

var running bool;  // global variable if client is running

var debug = flag.Bool("d", false, "set the debug modus( print informations )")

// func Log(v ...): loging. give log information if debug is true
func Log(v ...) {
    if *debug == true {
        ret := fmt.Sprint(v);
        log.Stdoutf("CLIENT: %s", ret);
    }
}

// func test(): testing for error
func test(err os.Error, mesg string) {
    if err!=nil {
        log.Stderr("CLIENT: ERROR: ", mesg);
         os.Exit(-1);
    } else
        Log("Ok: ", mesg);
}

// read from connection and return true if ok
func Read(con *net.Conn) string{
    var buf [4048]byte;  ///////////////////////
    _, err := con.Read(&buf); ///////////////////////
    if err!=nil {
        con.Close();
        running=false;
        return "Error in reading!";
    }
    str := string(&buf);  ///////////////////////
    fmt.Println();
    return string(str);
}

// clientsender(): read from stdin and send it via network
func clientsender(cn *net.Conn) {
    reader := bufio.NewReader(os.Stdin);
    for {
        fmt.Print("you> ");
        input, _ := reader.ReadBytes('\n');
        if bytes.Equal(input, strings.Bytes("/quit\n")) {
            cn.Write(strings.Bytes("/quit"));
            running = false;
            break;
        }
        Log("clientsender(): send: ", string(input[0:len(input)-1]));
        cn.Write(input[0:len(input)-1]);/////////////////////////
    }
}

// clientreceiver(): wait for input from network and print it out
func clientreceiver(cn *net.Conn) {
    for running {///////////////////////////////
        fmt.Println(Read(cn));
        fmt.Print("you> ");
    }
}

func main() {
    flag.Parse();
    running = true;
    Log("main(): start ");
    
    // connect
    destination := "127.0.0.1:9988";
    Log("main(): connecto to ", destination);
    cn, err := net.Dial("tcp", "", destination);
    test(err, "dialing");
    defer cn.Close();
    Log("main(): connected ");

    // get the user name
    fmt.Print("Please give you name: ");
    reader := bufio.NewReader(os.Stdin);
    name, _ := reader.ReadBytes('\n');

    //cn.Write(strings.Bytes("User: "));
    cn.Write(name[0:len(name)-1]);

    // start receiver and sender
    Log("main(): start receiver");
    go clientreceiver(&cn);
    Log("main(): start sender");
    go clientsender(&cn);
    
    // wait for quiting (/quit). run until running is true
    for ;running; {
        time.Sleep(1*1e9);
    }
    Log("main(): stoped");
}
