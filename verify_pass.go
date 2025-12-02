package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "password123"
	// Generate
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Generated Hash: %s\n", string(hash))

	// Verify
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		fmt.Printf("Verification Failed: %v\n", err)
	} else {
		fmt.Println("Verification Successful!")
	}
}
