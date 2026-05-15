package auth

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "strings"
)

func HashPassword(password string) (string, error) {
    if password == "" {
        return "", errors.New("empty password")
    }

    salt := make([]byte, 16)
    if _, err := rand.Read(salt); err != nil {
        return "", err
    }

    sum := sha256.Sum256(append(salt, []byte(password)...))
    return hex.EncodeToString(salt) + ":" + hex.EncodeToString(sum[:]), nil
}

func CheckPassword(hash, password string) error {
    parts := strings.Split(hash, ":")
    if len(parts) != 2 {
        return errors.New("invalid password hash")
    }
    salt, err := hex.DecodeString(parts[0])
    if err != nil {
        return err
    }
    expected, err := hex.DecodeString(parts[1])
    if err != nil {
        return err
    }
    sum := sha256.Sum256(append(salt, []byte(password)...))
    if hex.EncodeToString(sum[:]) != hex.EncodeToString(expected) {
        return errors.New("password mismatch")
    }
    return nil
}
