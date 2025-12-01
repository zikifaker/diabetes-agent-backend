package knowledgebase

import (
	"crypto/hmac"
	"crypto/sha256"
	"diabetes-agent-backend/config"
	"diabetes-agent-backend/response"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"time"

	"github.com/aliyun/credentials-go/credentials"
)

const (
	product = "oss"

	// 访问OSS的凭证有效期（秒）
	roleSessionExpiration = 3600
)

var (
	bucketName = config.Cfg.OSS.BucketName
	region     = config.Cfg.OSS.Region
)

// GeneratePolicyToken 获取前端上传文件到OSS的凭证
func GeneratePolicyToken(email string) (*response.GetPolicyTokenResponse, error) {
	host := fmt.Sprintf("https://%s.oss-%s.aliyuncs.com", bucketName, region)

	// 文件前缀
	dir := email + "/"

	config := new(credentials.Config).
		SetType("ram_role_arn").
		SetAccessKeyId(config.Cfg.OSS.AccessKeyID).
		SetAccessKeySecret(config.Cfg.OSS.AccessKeySecret).
		SetRoleArn(config.Cfg.OSS.RoleARN).
		SetRoleSessionName("Role_Session_Name").
		SetPolicy("").
		SetRoleSessionExpiration(roleSessionExpiration)

	// 创建凭证提供器
	provider, err := credentials.NewCredential(config)
	if err != nil {
		return nil, fmt.Errorf("fail to create credential: %v", err)
	}

	// 获取凭证
	cred, err := provider.GetCredential()
	if err != nil {
		return nil, fmt.Errorf("fail to get credential: %v", err)
	}

	// 构建policy
	utcTime := time.Now().UTC()
	date := utcTime.Format("20060102")
	expiration := utcTime.Add(1 * time.Hour)
	policyMap := map[string]any{
		"expiration": expiration.Format("2006-01-02T15:04:05.000Z"),
		"conditions": []any{
			map[string]string{"bucket": bucketName},
			map[string]string{"x-oss-signature-version": "OSS4-HMAC-SHA256"},
			map[string]string{"x-oss-credential": fmt.Sprintf("%v/%v/%v/%v/aliyun_v4_request", *cred.AccessKeyId, date, region, product)},
			map[string]string{"x-oss-date": utcTime.Format("20060102T150405Z")},
			map[string]string{"x-oss-security-token": *cred.SecurityToken},
		},
	}

	policy, err := json.Marshal(policyMap)
	if err != nil {
		return nil, fmt.Errorf("fail to marshal policy: %v", err)
	}

	// 构造待签名字符串
	stringToSign := base64.StdEncoding.EncodeToString(policy)

	signature := generateSignature(stringToSign, cred, date)

	policyToken := &response.GetPolicyTokenResponse{
		Policy:           stringToSign,
		SecurityToken:    *cred.SecurityToken,
		SignatureVersion: "OSS4-HMAC-SHA256",
		Credential:       fmt.Sprintf("%v/%v/%v/%v/aliyun_v4_request", *cred.AccessKeyId, date, region, product),
		Date:             utcTime.UTC().Format("20060102T150405Z"),
		Signature:        signature,
		Host:             host,
		Dir:              dir,
	}

	return policyToken, nil
}

func generateSignature(stringToSign string, cred *credentials.CredentialModel, date string) string {
	hmacHash := func() hash.Hash {
		return sha256.New()
	}

	signingKey := "aliyun_v4" + *cred.AccessKeySecret

	h1 := hmac.New(hmacHash, []byte(signingKey))
	io.WriteString(h1, date)
	h1Key := h1.Sum(nil)

	h2 := hmac.New(hmacHash, h1Key)
	io.WriteString(h2, region)
	h2Key := h2.Sum(nil)

	h3 := hmac.New(hmacHash, h2Key)
	io.WriteString(h3, product)
	h3Key := h3.Sum(nil)

	h4 := hmac.New(hmacHash, h3Key)
	io.WriteString(h4, "aliyun_v4_request")
	h4Key := h4.Sum(nil)

	h := hmac.New(hmacHash, h4Key)
	io.WriteString(h, stringToSign)
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}
