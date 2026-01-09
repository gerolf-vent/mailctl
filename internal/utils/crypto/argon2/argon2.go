package argon2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrMismatchedHashAndPassword = errors.New("hashedPassword is not the hash of the given password")
)

type InvalidHashFormatError string

func (ife InvalidHashFormatError) Error() string {
	return fmt.Sprintf("invalid argon2id hash format: %s", string(ife))
}

type InvalidHashParameterError string

func (ipe InvalidHashParameterError) Error() string {
	return fmt.Sprintf("invalid argon2id hash parameter: %s", string(ipe))
}

func parseArgon2IDHash(hashedPassword []byte) (version uint8, time, memory uint32, threads uint8, salt, hash []byte, err error) {
	if len(hashedPassword) == 0 {
		err = errors.New("hashed password is empty")
		return
	}

	hashParts := strings.Split(string(hashedPassword), "$")
	if len(hashParts) < 5 || len(hashParts) > 6 {
		err = InvalidHashFormatError("expected 5 to 6 parts in hash separated by '$' (first part must be empty)")
		return
	}

	var emptyPart, identifierPart, versionPart, parametersPart, saltPart, hashPart string
	if len(hashParts) == 6 {
		emptyPart = hashParts[0]
		identifierPart = hashParts[1]
		versionPart = hashParts[2]
		parametersPart = hashParts[3]
		saltPart = hashParts[4]
		hashPart = hashParts[5]
	} else {
		// Old format without version part
		emptyPart = hashParts[0]
		identifierPart = hashParts[1]
		versionPart = ""
		parametersPart = hashParts[2]
		saltPart = hashParts[3]
		hashPart = hashParts[4]
	}

	// Check if first part is empty (that implies that $ is the first character)
	if len(emptyPart) > 0 {
		err = InvalidHashFormatError("argon2id hash must start with a '$'")
		return
	}

	// Check identifier
	switch identifierPart {
	case "argon2i", "argon2d":
		err = InvalidHashFormatError("only argon2id is supported")
		return
	case "argon2id":
		// valid
	default:
		err = InvalidHashFormatError("expected argon2id identifier")
		return
	}

	// Parse version
	version = 10 // Fallback to default Argon2 version
	if len(versionPart) > 0 {
		var v int
		_, err = fmt.Sscanf(versionPart, "v=%d", &v)
		if err != nil {
			err = InvalidHashFormatError("unable to parse version")
			return
		}
		version = uint8(v)
	}

	// Check if version is supported
	if version != argon2.Version {
		err = InvalidHashParameterError(fmt.Sprintf("argon2 hash version %d is different from supported version %d", version, argon2.Version))
		return
	}

	// Parse parameters
	parameters := strings.Split(parametersPart, ",")
	if len(parameters) != 3 {
		err = InvalidHashFormatError("invalid parameters format")
		return
	}

	for _, paramStr := range parameters {
		paramParts := strings.SplitN(paramStr, "=", 2)
		if len(paramParts) != 2 {
			err = InvalidHashFormatError("invalid parameter format")
			return
		}

		key, value := paramParts[0], paramParts[1]

		switch key {
		case "m":
			var parsedValue int64
			parsedValue, err = strconv.ParseInt(value, 10, 32)
			if err != nil {
				err = InvalidHashParameterError("unable to parse memory parameter")
				return
			}
			memory = uint32(parsedValue)
		case "t":
			var parsedValue int64
			parsedValue, err = strconv.ParseInt(value, 10, 32)
			if err != nil {
				err = InvalidHashParameterError("unable to parse time parameter")
				return
			}
			time = uint32(parsedValue)
		case "p":
			var parsedValue int64
			parsedValue, err = strconv.ParseInt(value, 10, 32)
			if err != nil {
				err = InvalidHashParameterError("unable to parse threads parameter")
				return
			}
			threads = uint8(parsedValue)
		default:
			err = InvalidHashParameterError("unknown parameter")
			return
		}
	}

	// Check if all parameters are set
	if memory == 0 {
		err = InvalidHashParameterError("memory parameter not set")
		return
	}
	if time == 0 {
		err = InvalidHashParameterError("time parameter not set")
		return
	}
	if threads == 0 {
		err = InvalidHashParameterError("threads parameter not set")
		return
	}

	// Decode salt
	salt, err = base64.RawStdEncoding.DecodeString(saltPart)
	if err != nil {
		err = InvalidHashParameterError("unable to decode salt")
		return
	}

	// Decode hash
	hash, err = base64.RawStdEncoding.DecodeString(hashPart)
	if err != nil {
		err = InvalidHashParameterError("unable to decode hash")
		return
	}

	return
}

func CompareHashAndPassword(hashedPassword, password []byte) error {
	version, time, memory, threads, salt, referenceHash, err := parseArgon2IDHash(hashedPassword)
	if err != nil {
		return err
	}

	if version > argon2.Version {
		return InvalidHashParameterError("unsupported argon2 version")
	}

	computedHash := argon2.IDKey(password, salt, time, memory, threads, uint32(len(referenceHash)))

	if subtle.ConstantTimeCompare(computedHash, referenceHash) != 1 {
		return ErrMismatchedHashAndPassword
	}

	return nil
}

func GenerateFromPassword(password []byte, time, memory uint32, threads uint8, keyLen uint32) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	hash := argon2.IDKey(password, salt, time, memory, threads, keyLen)

	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hash)

	hashedPassword := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, memory, time, threads, saltBase64, hashBase64)
	return []byte(hashedPassword), nil
}
