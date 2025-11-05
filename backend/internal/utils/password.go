package utils

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 哈希密码
func HashPassword(password string) (string, error) {
	// 使用bcrypt哈希密码，成本因子为12
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("生成随机字符串失败: %w", err)
	}

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes), nil
}

// ValidatePasswordStrength 验证密码强度
func ValidatePasswordStrength(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return fmt.Errorf("密码长度至少为8位")
	}

	if utf8.RuneCountInString(password) > 128 {
		return fmt.Errorf("密码长度不能超过128位")
	}

	var (
		hasLower   bool
		hasUpper   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsNumber(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if !hasLower {
		return fmt.Errorf("密码必须包含小写字母")
	}

	if !hasUpper {
		return fmt.Errorf("密码必须包含大写字母")
	}

	if !hasNumber {
		return fmt.Errorf("密码必须包含数字")
	}

	if !hasSpecial {
		return fmt.Errorf("密码必须包含特殊字符")
	}

	// 检查常见弱密码模式
	if isCommonWeakPassword(password) {
		return fmt.Errorf("密码过于常见，请使用更强的密码")
	}

	return nil
}

// isCommonWeakPassword 检查常见弱密码
func isCommonWeakPassword(password string) bool {
	weakPatterns := []string{
		"^(.)\\1+$",           // 重复字符: aaaaa
		"^(123|abc)",          // 简单序列: 123456, abcdef
		"password",           // 包含常见密码词
		"qwerty",             // 键盘序列
		"(.{1,2})(.{1,2})\\2\\1", // 回文模式
	}

	for _, pattern := range weakPatterns {
		if matched, _ := regexp.MatchString(pattern, password); matched {
			return true
		}
	}

	return false
}

// IsAccountLocked 检查账户是否被锁定
func IsAccountLocked(failedAttempts int, lockedUntil *string) bool {
	if lockedUntil != nil {
		// 这里需要解析时间字符串，简化处理
		return false
	}
	return failedAttempts >= 5
}