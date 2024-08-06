package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// Encryption function
func encrypt(plaintext []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Ensure the plaintext length is a multiple of the block size
	if len(plaintext)%aes.BlockSize != 0 {
		panic("plaintext is not a multiple of the block size")
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	copy(ciphertext[:aes.BlockSize], iv)

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// Decryption function
func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Ensure the ciphertext length is a multiple of the block size
	if len(ciphertext)%aes.BlockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	return ciphertext, nil
}

// BEAST attack simulation
func beastAttack(ciphertext []byte, knownPlaintext []byte, targetBlockIndex int) []byte {
	iv := ciphertext[:aes.BlockSize]
	targetBlock := ciphertext[aes.BlockSize*targetBlockIndex : aes.BlockSize*(targetBlockIndex+1)]

	fakeBlock := make([]byte, aes.BlockSize)
	for i := 0; i < aes.BlockSize; i++ {
		fakeBlock[i] = iv[i] ^ targetBlock[i] ^ knownPlaintext[i]
	}

	return fakeBlock
}

// Generate random IV
func generateRandomIV() ([]byte, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	return iv, nil
}

func main() {
	key, _ := hex.DecodeString("6368616e676520746869732070617373")
	plaintext := []byte("This is a secret message for CBC mode demo!")

	// Use random IV
	iv, err := generateRandomIV()
	if err != nil {
		fmt.Println("Error generating IV:", err)
		return
	}

	fmt.Println("Original plaintext:", string(plaintext))

	// Encrypt
	ciphertext, err := encrypt(plaintext, key, iv)
	if err != nil {
		fmt.Println("Error encrypting:", err)
		return
	}
	fmt.Printf("Encrypted ciphertext (hex): %x\n", ciphertext)

	// Decrypt
	decrypted, err := decrypt(ciphertext, key)
	if err != nil {
		fmt.Println("Error decrypting:", err)
		return
	}
	fmt.Println("Decrypted plaintext:", string(decrypted))

	// Simulate BEAST attack
	fmt.Println("\nSimulating BEAST attack...")
	knownPlaintext := []byte("This is a secret")
	targetBlockIndex := 1
	fakeBlock := beastAttack(ciphertext, knownPlaintext, targetBlockIndex)

	// Construct attacked ciphertext
	attackedCiphertext := append(fakeBlock, ciphertext[aes.BlockSize*(targetBlockIndex+1):]...)

	// Decrypt attacked ciphertext
	attackedDecrypted, _ := decrypt(attackedCiphertext, key)
	fmt.Println("Recovered plaintext from attack:", string(attackedDecrypted[:aes.BlockSize]))

	// Demonstrate defense with random IV
	fmt.Println("\nDefending with random IV...")
	newIV, _ := generateRandomIV()
	defendedCiphertext, _ := encrypt(plaintext, key, newIV)
	fmt.Printf("Ciphertext with random IV (hex): %x\n", defendedCiphertext)
	defendedDecrypted, _ := decrypt(defendedCiphertext, key)
	fmt.Println("Decrypted plaintext with random IV:", string(defendedDecrypted))
}
