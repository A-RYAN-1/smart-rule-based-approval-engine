package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
    fmt.Println(string(hash))

    hash2, _ := bcrypt.GenerateFromPassword([]byte("manager123"), bcrypt.DefaultCost)
    fmt.Println(string(hash2))
}