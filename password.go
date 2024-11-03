package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// paraphrased from https://www.alexedwards.net/blog/how-to-hash-and-verify-passwords-with-argon2-in-go

type PasswordParams struct {
	Memory uint32
	Iterations uint32
	Parallelism uint8
	SaltLength uint32
	KeyLength uint32
}

func DefaultParams() *PasswordParams {
	return &PasswordParams{
		Memory: 19456,
		Iterations: 2,
		Parallelism: 1,

		SaltLength: 16,
		KeyLength: 32,
	}
}

func genRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func HashPassword(password string, p *PasswordParams) (string, error) {
	salt, err := genRandomBytes(p.SaltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.Memory,
		p.Iterations,
		p.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

func DecodePasswordHash(encodedHash string) (p *PasswordParams, salt, hash []byte, err error) {
	tokens := strings.Split(encodedHash, "$")
	if len(tokens) != 6 {
		return nil, nil, nil, err
	}

	if tokens[1] != "argon2id" {
		return nil, nil, nil, errors.New("Invalid hash method")
	}

	var version int
	_, err = fmt.Sscanf(tokens[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, errors.New("Incompatible version")
	}

	p = &PasswordParams{}

	_, err = fmt.Sscanf(tokens[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(tokens[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(tokens[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}

func ComparePasswordAndHash(password string, encodedHash string) (bool, error) {
	p, salt, hash, err := DecodePasswordHash(encodedHash)
	if err != nil {
		return false, err
	}

	fmt.Printf("%x\n", hash)

	testHash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)
	fmt.Printf("%x\n", testHash)

	return subtle.ConstantTimeCompare(hash, testHash) == 1, nil
}
