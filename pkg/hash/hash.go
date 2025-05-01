package hash

import "fmt"

type Hasher interface {
	Hash(password string) (string, error)
	Compare(password string, hash string) (bool, error)
}

type Argon2Config struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type Config struct {
	Algo string
	Argon2Config
}

func New(config Config) (Hasher, error) {
	switch config.Algo {
	case "argon2":
		return &ArgonHasher{config: config.Argon2Config}, nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", config.Algo)
	}
}
