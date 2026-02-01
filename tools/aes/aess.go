package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// AES-CBC 解密
func DecryContentWithAESCBC(ciphertext []byte, key []byte, iv []byte) ([]byte, error) {
	if l := len(key); l != 16 && l != 24 && l != 32 {
		return nil, errors.New("key length must be 16, 24 or 32 bytes")
	}
	if len(iv) != aes.BlockSize { // 检查 IV 长度
		return nil, errors.New("IV must be 16 bytes for CBC")
	}

	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 { // 检查密文长度
		return nil, errors.New("ciphertext length is not multiple of block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	return pkcs7Unpad(plaintext)
}

// AES-CBC 加密
func EncryContentWithAESCBC(plaintext []byte, key []byte, iv []byte) ([]byte, error) {
	if l := len(key); l != 16 && l != 24 && l != 32 {
		return nil, errors.New("key length must be 16, 24 or 32 bytes")
	}
	if len(iv) != aes.BlockSize { // 检查 IV 长度
		return nil, errors.New("IV must be 16 bytes for CBC")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext = pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(plaintext))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	return ciphertext, nil
}

// AES-GCM 解密
func DecryContentWithAESGCM(ciphertext []byte, key []byte, iv []byte) ([]byte, error) {
	if l := len(key); l != 16 && l != 24 && l != 32 {
		return nil, errors.New("key length must be 16, 24 or 32 bytes")
	}
	if len(iv) != 12 { // 检查 nonce 长度
		return nil, errors.New("nonce must be 12 bytes for GCM")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		return nil, err
	}

	decrypted, err := aesGCM.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}

// AES-GCM 加密
func EncryContentWithAESGCM(plaintext []byte, key []byte, iv []byte) ([]byte, error) {
	if l := len(key); l != 16 && l != 24 && l != 32 {
		return nil, errors.New("key length must be 16, 24 or 32 bytes")
	}
	if len(iv) != 12 { // 检查 nonce 长度
		return nil, errors.New("nonce must be 12 bytes for GCM")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nil, iv, plaintext, nil)
	return ciphertext, nil
}

// pkcs7Pad PKCS7填充
func pkcs7Pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

// pkcs7Unpad PKCS7去除填充（带验证）
func pkcs7Unpad(src []byte) ([]byte, error) {
	length := len(src)
	if length == 0 {
		return nil, errors.New("invalid PKCS7 padding")
	}

	unpadding := int(src[length-1])
	if unpadding > length || unpadding <= 0 {
		return nil, errors.New("invalid PKCS7 padding")
	}

	// 验证所有填充字节是否一致
	for i := length - unpadding; i < length; i++ {
		if src[i] != byte(unpadding) {
			return nil, errors.New("invalid PKCS7 padding")
		}
	}

	return src[:length-unpadding], nil
}
