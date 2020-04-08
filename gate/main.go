package main

import (
    "log"
    "bufio"
    "os"

    "easyserver/lib/gate"
)

func main() {
    log.Println("start gate...")

    gate.Open("127.0.0.1:44000")

    log.Println("input anykey to exit...")
    input := bufio.NewReader(os.Stdin)
    input.ReadString('\n')
}