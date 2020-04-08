package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

func main() {
	var preimage [32]byte
	if _, err := rand.Read(preimage[:]); err != nil {
		return
	}
	hash := sha256.Sum256(preimage[:])
	fmt.Printf("preimage: %x\nhash:     %x\n", preimage, hash)
}
