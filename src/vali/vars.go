package vali


import (
    "log"
    "sync"
)


var (
    Logger       = new(log.Logger)
    handleWG     sync.WaitGroup
    RLIMIT_NPROC = 0x6
)