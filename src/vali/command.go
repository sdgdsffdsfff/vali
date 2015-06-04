package vali


import (
    "os"
    "os/exec"
    "syscall"
    "time"
)


type ProcessManager struct {
    Command   *exec.Cmd
    StartTime int64
    Timeout   int64
    Exit      chan bool
}


type Result struct {
    Code      int    `json:"code"`
    Info      string `json:"info"`
    TimeStamp int64  `json:"timestamp"`
}


func (pm *ProcessManager) RunCommand() *Result {
    ret := &Result{}
    err := os.Chmod(pm.Command.Args[0], 0755)
    if os.IsNotExist(err) {
        ret.Code = -9
        ret.Info = err.Error()
        return ret
    }
    err = os.MkdirAll(pm.Command.Dir, 0755)
    if err != nil {
        ret.Code = -9
        ret.Info = err.Error()
        return ret
    } else {
        uid, gid := pm.Command.SysProcAttr.Credential.Uid, pm.Command.SysProcAttr.Credential.Gid
        os.Chown(pm.Command.Dir, int(uid), int(gid))
    }
    pm.Command.SysProcAttr.Setpgid = true
    if err = pm.Command.Start(); err != nil {
        ret.Code = -9
        ret.Info = err.Error()
        return ret
    }
    defer func() { pm.Exit <- true }()
    go pm.CheckTimeout()
    if stat := pm.Command.Wait(); stat != nil {
        ret.Code = pm.Command.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
        Logger.Printf("command: %v error: %s", pm.Command.Args, stat.Error())
        ret.Info = stat.Error()
    } else {
        ret.Code = 0
    }
    ret.TimeStamp = time.Now().Unix()
    return ret
}


func (pm *ProcessManager) CheckTimeout() {
    process := pm.Command.Process
    for {
        select {
        case <-pm.Exit:
            return
        default:
        }
        current := time.Now().Unix()
        if duration := current - pm.StartTime; duration > pm.Timeout {
            // pgroup kill, kill process and all its children.
            syscall.Kill(-process.Pid, 9)
            Logger.Printf("command: %v timeout, killed", pm.Command.Args)
        }
        time.Sleep(time.Duration(1) * time.Second)
    }
}