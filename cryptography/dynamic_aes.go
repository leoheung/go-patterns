package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
)

/*
DynamicAES: 动态 AES 密钥管理与加密工具

简介:
DynamicAES 是一个用于动态生成、传输和管理 AES 密钥的工具，结合了 ECDH 密钥协商和 AES-GCM 加密。
它适用于需要安全传递对称密钥的场景，例如客户端与服务器之间的通信。

核心功能:
1. 动态生成 AES 密钥。
2. 使用 ECDH (X25519) 协商共享密钥，安全传递 AES 密钥。
3. 提供 AES-GCM 加密和解密功能，支持任意对象的序列化加密。

使用场景:
- 客户端生成随机 AES 密钥，通过临时 ECDH 密钥加密后发送给服务器。
- 服务器解密并设置 AES 密钥，用于后续的业务数据加密和解密。
- 适合短期会话密钥传递和数据保护。

安全性:
- 使用 X25519 (ECDH) 协商共享密钥，确保密钥传输的安全性。
- 使用 AES-GCM 提供认证加密，防止数据篡改。
- 密钥和数据均以 Base64 编码传输，便于集成。

示例代码:
1. 客户端生成 AES 密钥并发送加密包给服务器。
2. 服务器解密包并设置 AES 密钥。
3. 双方使用 AES 密钥加密和解密业务数据。

Usage Example:
// 1. 初始化 (App端)
client := NewDynamicAES(32)

// 2. 握手 (App端加密自己的密钥，发给服务器)
// pubKeyFromApp 是你之前生成的 Base64 ECC 公钥
encryptedPackage, _ := client.GetKey(pubStr)

// 3. 业务加密
type GameData struct { Gold int; Pos string }
cipherText, _ := client.Encrypt(GameData{Gold: 99, Pos: "10,20"})
fmt.Println("\n发给服务器的数据:", cipherText)


// 4. 服务器端设置AES key
server := NewDynamicAES(32)
server.SetKey(privStr,encryptedPackage)


var encryptedData  GameData
server.Decrypt(cipherText, &encryptedData)
fmt.Println("服务器收到:", encryptedData.Gold, encryptedData.Pos)

*/

type DynamicAES struct {
	AESKey []byte // 内部使用 byte 数组处理更高效
}

// NewDynamicAES 初始化
func NewDynamicAES(length int) *DynamicAES {
	if length <= 0 {
		return &DynamicAES{}
	}
	key := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil
	}
	return &DynamicAES{AESKey: key}
}

// GetKey 使用公钥加密当前的 AES 密钥 (用于 App 端发送给服务器)
// pk 为 Base64 格式的 PKIX 公钥
func (d *DynamicAES) GetKey(pk string) (string, error) {
	// 1. 解析公钥
	pubBytes, _ := base64.StdEncoding.DecodeString(pk)
	pubInterface, err := x509.ParsePKIXPublicKey(pubBytes)
	if err != nil {
		return "", err
	}
	pub := pubInterface.(*ecdh.PublicKey)

	// 2. ECC 无法直接像 RSA 那样加密，标准做法是生成临时密钥对进行 ECDH 协商
	// 或者使用更高级的 ECIES 封装。为了简单直观且符合您的设计：
	// 我们这里模拟“密封箱”逻辑：生成临时密钥并计算共享密钥来保护 AES Key
	ephemeralPriv, _ := ecdh.X25519().GenerateKey(rand.Reader)
	sharedSecret, _ := ephemeralPriv.ECDH(pub)

	// 使用共享密钥加密真实的 AES Key
	encryptedKey, _ := aesEncrypt(sharedSecret[:32], d.AESKey)

	// 返回：临时公钥 + 加密后的 AES Key (组合成一个包发给服务器)
	result := append(ephemeralPriv.PublicKey().Bytes(), encryptedKey...)
	return base64.StdEncoding.EncodeToString(result), nil
}

// SetKey 服务器端使用私钥解开包并设置密钥
// sk 为 Base64 格式的 PKCS8 私钥
func (d *DynamicAES) SetKey(sk string, encryptedPackage string) error {
	// 1. 解析私钥
	privBytes, _ := base64.StdEncoding.DecodeString(sk)
	privInterface, err := x509.ParsePKCS8PrivateKey(privBytes)
	if err != nil {
		return err
	}
	priv := privInterface.(*ecdh.PrivateKey)

	// 2. 解析数据包 (前 32 字节是临时公钥)
	packageBytes, _ := base64.StdEncoding.DecodeString(encryptedPackage)
	if len(packageBytes) < 32 {
		return errors.New("invalid package")
	}
	remotePubBytes := packageBytes[:32]
	cipherData := packageBytes[32:]

	remotePub, _ := ecdh.X25519().NewPublicKey(remotePubBytes)
	sharedSecret, _ := priv.ECDH(remotePub)

	// 3. 解密出真正的 AES Key
	realKey, err := aesDecrypt(sharedSecret[:32], cipherData)
	if err != nil {
		return err
	}
	d.AESKey = realKey
	return nil
}

// Encrypt 加密任意对象为 Base64 字符串
func (d *DynamicAES) Encrypt(obj any) (string, error) {
	plaintext, _ := json.Marshal(obj)
	ciphertext, err := aesEncrypt(d.AESKey, plaintext)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 Base64 字符串到对象
func (d *DynamicAES) Decrypt(cipherTextStr string, obj any) error {
	ciphertext, _ := base64.StdEncoding.DecodeString(cipherTextStr)
	plaintext, err := aesDecrypt(d.AESKey, ciphertext)
	if err != nil {
		return err
	}
	return json.Unmarshal(plaintext, obj)
}

// --- 内部工具函数：AES-GCM 标准实现 ---

func aesEncrypt(key, plaintext []byte) ([]byte, error) {
	// 确保 Key 是 32 字节 (AES-256)
	hasher := sha256.Sum256(key)
	block, _ := aes.NewCipher(hasher[:])
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	// 返回 nonce + 密文
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func aesDecrypt(key, ciphertext []byte) ([]byte, error) {
	hasher := sha256.Sum256(key)
	block, _ := aes.NewCipher(hasher[:])
	gcm, _ := cipher.NewGCM(block)
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("cipher too short")
	}
	nonce, actualCipher := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, actualCipher, nil)
}
