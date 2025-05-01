package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type ArgonHasher struct {
	config Argon2Config
}

var (
	ErrorInvalidHash         = errors.New("The encoded hash is not in valid format")
	ErrorIncompatibleVersion = errors.New("Incompatible version")
)

func (hasher *ArgonHasher) Hash(password string) (string, error) {
	salt, err := hasher.generateRandomBytes(hasher.config.SaltLength)
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey(
		[]byte(password), salt,
		hasher.config.Iterations,
		hasher.config.Memory,
		hasher.config.Parallelism,
		hasher.config.KeyLength)
	base64Salt := base64.RawStdEncoding.EncodeToString(salt)
	base64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encodePassword := fmt.Sprintf("$argon2id,$v=%d,$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		hasher.config.Memory,
		hasher.config.Iterations,
		hasher.config.Parallelism,
		base64Salt, base64Hash)
	return encodePassword, nil
}

func (hasher *ArgonHasher) Compare(password string, encodedHash string) (bool, error) {
	config, salt, hash, err := hasher.decodeHash(encodedHash)
	if err != nil {
		return false, nil
	}
	hashedPassword := argon2.IDKey(
		[]byte(password), salt,
		config.Iterations,
		config.Memory,
		config.Parallelism,
		config.KeyLength)

	// Check that the contents of the hashed passwords are identical.
	// Note:
	// We are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, hashedPassword) == 1 {
		return true, nil
	}
	return false, nil
}
func (hasher *ArgonHasher) decodeHash(encodePassword string) (*Argon2Config, []byte, []byte, error) {
	metadata := strings.Split(encodePassword, "$")
	if len(metadata) != 6 {
		return nil, nil, nil, ErrorInvalidHash
	}
	var version int
	_, err := fmt.Sscanf(metadata[2], "v=%d", version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrorIncompatibleVersion
	}
	config := &Argon2Config{}
	_, err = fmt.Sscanf("m=%d,t=%d,p=%d", metadata[3], config.Memory, config.Iterations, config.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(metadata[4])
	if err != nil {
		return nil, nil, nil, err
	}
	hash, err := base64.RawStdEncoding.Strict().DecodeString(metadata[5])
	if err != nil {
		return nil, nil, nil, err
	}
	config.KeyLength = uint32(len(hash))
	return config, salt, hash, nil
}
func (hasher *ArgonHasher) generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
