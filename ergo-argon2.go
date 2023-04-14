package ergoArgon2

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"

	"golang.org/x/crypto/argon2"
)

// Current default parameters, Argon2id version 0x13
const (
	timeCost    = 3
	memoryCost  = 64 * 1024
	parallelism = 4
	keyLength   = 32
	saltLength  = 16
	version     = 0x13
)

type Hash struct {
	salt        []byte
	hash        []byte
	timeCost    uint32
	memoryCost  uint32
	parallelism uint8
}

func HashPassword(password string) string {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		panic(err)
	}
	ID := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, parallelism, keyLength)

	hash := Hash{
		salt:        []byte(base64.RawStdEncoding.EncodeToString(salt)),
		hash:        []byte(base64.RawStdEncoding.EncodeToString(ID)),
		timeCost:    timeCost,
		memoryCost:  memoryCost,
		parallelism: parallelism,
	}

	return hash.String()
}

func (h *Hash) String() string {
	return fmt.Sprintf("$argon2id$v=%02d$m=%02d,t=%02d,p=%02d$%s$%s", version, h.memoryCost, h.timeCost, h.parallelism, h.salt, h.hash)
}

func parseHash(hash string) (Hash, error) {
	// parse phc-string-formatted hash
	// $argon2id$v=19$m=65536,t=3,p=4$YXNkZmFzZGxmbnNkYWZoYXNkZg$YXNkZmFzZGxmbnNkYWZoYXNkZg
	// split on '$'
	parts := bytes.Split([]byte(hash), []byte{'$'})
	invalidHashErr := errors.New("invalid hash")
	if len(parts) != 6 {
		return Hash{}, invalidHashErr
	}
	// check algorithm
	if string(parts[1]) != "argon2id" {
		return Hash{}, invalidHashErr
	}
	// check version
	if string(parts[2]) != "v=19" {
		return Hash{}, invalidHashErr
	}
	// split parameters
	params := bytes.Split(parts[3], []byte{','})
	if len(params) != 3 {
		return Hash{}, invalidHashErr
	}
	// get m, t, p
	memoryCost, err := strconv.ParseUint(string(params[0][2:]), 10, 32)
	if err != nil {
		return Hash{}, invalidHashErr
	}
	timeCost, err := strconv.ParseUint(string(params[1][2:]), 10, 32)
	if err != nil {
		return Hash{}, invalidHashErr
	}
	parallelism, err := strconv.ParseUint(string(params[2][2:]), 10, 8)
	if err != nil {
		return Hash{}, invalidHashErr
	}
	return Hash{
		salt:        parts[4],
		hash:        parts[5],
		timeCost:    uint32(timeCost),
		memoryCost:  uint32(memoryCost),
		parallelism: uint8(parallelism),
	}, nil
}

func VerifyPassword(password string, hash string) bool {
	// Parse hash
	h, err := parseHash(hash)
	if err != nil {
		return false
	}

	// Get salt
	salt, err := base64.RawStdEncoding.DecodeString(string(h.salt))
	if err != nil {
		return false
	}

	// Get hash
	ID, err := base64.RawStdEncoding.DecodeString(string(h.hash))
	if err != nil {
		return false
	}

	newID := argon2.IDKey([]byte(password), salt, h.timeCost, h.memoryCost, h.parallelism, keyLength)

	return bytes.Equal(ID, newID)
}