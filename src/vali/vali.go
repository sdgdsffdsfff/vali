package vali

import (
	"bytes"
	"fmt"
	"io"
	"logme"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Reply struct {
	Type      string `json:"type"`
	Msg       string `json:"message"`
	TimeStamp int64  `json:"timestamp"`
}

type ConInfo struct {
	Connection net.Conn
	OutQueue   chan []byte
	Closed     chan bool
	handleWG   *sync.WaitGroup
}

func StartServer() {
	host, _ := os.Hostname()
	ips, _ := net.LookupIP(host)
	privateIP := ""
	for _, ip := range ips {
		tmp := ip.String()
		if strings.HasPrefix(tmp, "10.") {
			privateIP = tmp
			break
		}
	}

	if privateIP == "" {
		fmt.Println("can not find private ip")
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("%s:1022", privateIP))
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	// set process based max process number and max open files
	proc_limit := syscall.Rlimit{Cur: 65535, Max: 65535}
	err = syscall.Setrlimit(RLIMIT_NPROC, &proc_limit)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	nofile_limit := syscall.Rlimit{Cur: 655350, Max: 655350}
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &nofile_limit)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	conf := &logme.Config{
		LogFile:   "vali.log",
		SizeLimit: 10000000,
	}
	logObj, err := conf.InitLogger()
	if err != nil {
		fmt.Print(err)
	}
	defer logObj.Close()
	go logObj.AutoRoll()
	Logger = logObj.LogMe
	for {
		c, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		info := &ConInfo{
			Connection: c,
			OutQueue:   make(chan []byte),
			Closed:     make(chan bool),
			handleWG:   new(sync.WaitGroup),
		}
		Logger.Printf("accept conneciton from: %v success", c.RemoteAddr())
		go info.handleConnection()
	}
}

func (c *ConInfo) handleConnection() {
	defer c.Close()
	go c.sendConn()
	go c.readConn()
	select {
	case <-c.Closed:
	}
}

func (c *ConInfo) Close() {
	Logger.Printf("close connection for: %v success", c.Connection.RemoteAddr())
	c.Connection.Close()
	close(c.OutQueue)
}

func (c *ConInfo) readLength(l int) ([]byte, error) {

	var data []byte

	for {
		recvData := make([]byte, l)

		length, err := c.Connection.Read(recvData)
		if err != nil {
			return nil, err
		}
		/* If use telnet to send message, CRLF will be received.
		   Length will two characters more than specified.
		   Delete CRLF and continue to receive new message.
		*/
		if str := len(bytes.TrimSpace(recvData[0:length])); str == 0 {
			length = 0
			continue
		}

		data = append(data, recvData[0:length]...)
		if length == l {
			return data, nil
		} else if l > length {
			l -= length
		}
	}
}

func (c *ConInfo) readConn() {
	for {
		data, err := c.readLength(10)
		if err != nil {
			break
		}
		l, err := strconv.Atoi(string(data[0:]))
		if err != nil {
			Logger.Printf("received wrong length: %s from: %s",
				string(data[0:]), c.Connection.RemoteAddr())
			break
		}
		line, err := c.readLength(l)
		if err != nil {
			break
		}
		c.handleWG.Add(1)
		go c.handleRecv(line)
	}
	c.handleWG.Wait()
	c.Closed <- true
}

func (c *ConInfo) sendConn() {
	for {
		select {
		case data, ok := <-c.OutQueue:
			if !ok {
				return
			}
			_, err := c.Connection.Write(data)
			if err != nil {
				Logger.Printf("%v is closed", c.Connection.RemoteAddr())
				return
			}
		}
	}
}

func (c *ConInfo) cmdStdout(cmd []string, output *io.ReadCloser, t string) {
	defer c.handleWG.Done()

	var buffer bytes.Buffer
	var line string
	data := make([]byte, 512)
	sendMessage := func(_line string) {
		if len(_line) == 0 {
			return
		} else if len(line) >= 20480 {
			_line = _line[:20480]
		}
		message := Reply{
			Msg:       _line,
			Type:      t,
			TimeStamp: time.Now().Unix(),
		}
		c.OutQueue <- Json(message)
		Logger.Printf("* %s: %s", t, _line)
	}

	for {
		length, err := (*output).Read(data)
		if err != nil {
			break
		}

		if length == 0 {
			break
		}
		buffer.Write(data[0:length])
		for {
			line, err = buffer.ReadString(byte('\n'))
			if err == io.EOF {
				buffer.WriteString(line)
				break
			}
			sendMessage(line)
		}
	}
	line = buffer.String()
	if line != "<nil>" {
		sendMessage(line)
	}
}

func (c *ConInfo) handleRecv(line []byte) {
	defer c.handleWG.Done()

	m, err := DecodeMessage(line)
	if err != nil {
		Logger.Printf("wrong json data :%s received", line)
		return
	}
	command := exec.Command(m.Command, m.Argument...)
	/* dosent work on crossing compile
	u, err := user.Lookup(m.RunUser)
	if err != nil {
		fmt.Println(err.Error())
	}
	uid, _ := strconv.ParseUint(u.Uid, 10, 32)
	gid, _ := strconv.ParseUint(u.Gid, 10, 32)
	*/
	arg1 := []string{"-u", m.RunUser}
	arg2 := []string{"-g", m.RunUser}

	uidOutput, _ := exec.Command("/usr/bin/id", arg1...).Output()
	gidOutput, _ := exec.Command("/usr/bin/id", arg2...).Output()

	uid, _ := strconv.ParseUint(strings.TrimSpace(string(uidOutput)), 10, 32)
	gid, _ := strconv.ParseUint(strings.TrimSpace(string(gidOutput)), 10, 32)

	user, err := user.Lookup(m.RunUser)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	command.SysProcAttr = &syscall.SysProcAttr{}
	command.SysProcAttr.Credential = &syscall.Credential{
		Uid: uint32(uid),
		Gid: uint32(gid),
	}

	// setting command enviroment variables
	env := []string{}
	env = append(env, fmt.Sprintf("HOME=%s", user.HomeDir))
	env = append(env, fmt.Sprintf("LOGNAME=%s", user.Username))
	env = append(env, fmt.Sprintf("USER=%s", user.Username))
	env = append(env, fmt.Sprintf("PWD=%s", m.RunDir))
	env = append(env, fmt.Sprintf("PATH=%s", "/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin/op"))
	env = append(env, fmt.Sprintf("LANG=%s", "en_US.UTF-8"))
	env = append(env, fmt.Sprintf("LANGUAGE=%s", "en_US.UTF-8"))
	env = append(env, fmt.Sprintf("LC_ALL=%s", "en_US.UTF-8"))
	env = append(env, fmt.Sprintf("LC_CTYPE=%s", "zh_CN.UTF-8"))

	command.Env = env
	command.Dir = m.RunDir

	pm := &ProcessManager{
		Command:   command,
		StartTime: time.Now().Unix(),
		Timeout:   m.Timeout,
		Exit:      make(chan bool),
	}
	Logger.Printf("* command: %#v started\n", command.Args)
	stdout, _ := command.StdoutPipe()
	stderr, _ := command.StderrPipe()

	c.handleWG.Add(2)
	go c.cmdStdout(command.Args, &stdout, "INFO")
	go c.cmdStdout(command.Args, &stderr, "ERROR")

	ret := pm.RunCommand()

	Logger.Printf("* command: %v code: %d, info: %s\n", command.Args, ret.Code, ret.Info)
	Logger.Printf("* command: %v finished!\n", command.Args)
	c.OutQueue <- Json(ret)
}
