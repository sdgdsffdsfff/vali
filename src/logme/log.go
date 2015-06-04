package logme


import (
    "fmt"
    "log"
    "os"
    "path"
    "time"
)


type Config struct {
    LogFile   string
    SizeLimit int64
}


type LogObj struct {
    File  *os.File
    LogMe *log.Logger
    Limit int64
    Term  chan bool
}


func (c *Config) InitLogger() (*LogObj, error) {
    dir := path.Dir(c.LogFile)
    err := os.MkdirAll(dir, 0755)
    if err != nil {
        fmt.Println("init log dir failed: %s", err.Error())
        return nil, err
    }
    logFile, err := os.OpenFile(c.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    info := &LogObj{
        File:  logFile,
        LogMe: log.New(logFile, "", log.LstdFlags),
        Limit: c.SizeLimit,
        Term:  make(chan bool),
    }
    return info, err
}


func (l *LogObj) AutoRoll() {
    for {
        stat, _ := l.File.Stat()
        if size := stat.Size(); size >= l.Limit {
            l.LogMe.Printf("resize log file, currtent:%d, limitation:%d\n", size, l.Limit)
            err := l.File.Truncate(1000000)
            if err != nil {
                fmt.Print(err)
            }
        }
        select {
        case <-l.Term:
            return
        default:
        }
        time.Sleep(time.Duration(1) * time.Second)
    }
}


func (l *LogObj) Close() {
    close(l.Term)
    l.File.Close()
}