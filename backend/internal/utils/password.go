package utils

import (
	"math"
	"math/rand"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GenerateOTP generates a random N-digit OTP.
func GenerateOTP(length int) string {
	// rand.Seed(time.Now().UnixNano())
	rand.New(rand.NewSource(time.Now().UnixNano()))
	min := int(math.Pow10(length - 1))
	max := int(math.Pow10(length)) - 1
	otp := rand.Intn(max-min+1) + min
	return strconv.Itoa(otp)
}
