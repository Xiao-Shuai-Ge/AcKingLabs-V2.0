// Package randkit 安全随机数：验证码、邀请码一律使用 crypto/rand。
package randkit

import (
	"crypto/rand"
	"math/big"
)

// DigitCode n 位数字验证码（首位可为 0，返回定长字符串）
func DigitCode(n int) string {
	s := make([]byte, n)
	for i := range s {
		v, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			panic(err) // crypto/rand 失败属于不可恢复的系统故障
		}
		s[i] = byte('0' + v.Int64())
	}
	return string(s)
}

// LetterCode n 位大写字母邀请码
func LetterCode(n int) string {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ" // 去掉易混淆的 I O
	s := make([]byte, n)
	for i := range s {
		v, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		s[i] = letters[v.Int64()]
	}
	return string(s)
}
