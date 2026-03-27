package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

type PresignParams struct {
	Key             string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Bucket          string
	SigningHost     string
	CDNHost         string
	ExpiresIn       int
}

func CreatePresignedURL(p PresignParams) string {
	now := time.Now().UTC()
	date := now.Format("20060102")
	datetime := now.Format("20060102T150405Z")

	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", date, p.Region)
	credential := fmt.Sprintf("%s/%s", p.AccessKeyID, credentialScope)

	queryParams := map[string]string{
		"X-Amz-Algorithm":     "AWS4-HMAC-SHA256",
		"X-Amz-Credential":    credential,
		"X-Amz-Date":          datetime,
		"X-Amz-Expires":       fmt.Sprintf("%d", p.ExpiresIn),
		"X-Amz-SignedHeaders": "host",
	}

	canonicalQueryString := buildCanonicalQueryString(queryParams)

	canonicalPath := "/" + p.Bucket + "/" + strings.TrimPrefix(p.Key, "/")

	canonicalHeaders := fmt.Sprintf("host:%s\n", p.SigningHost)
	signedHeaders := "host"

	canonicalRequest := strings.Join([]string{
		"GET",
		canonicalPath,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		"UNSIGNED-PAYLOAD",
	}, "\n")

	hashedCanonicalRequest := sha256hex(canonicalRequest)

	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		datetime,
		credentialScope,
		hashedCanonicalRequest,
	}, "\n")

	signingKey := deriveSigningKey(p.SecretAccessKey, date, p.Region)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	finalQuery := canonicalQueryString + "&X-Amz-Signature=" + signature

	cdnHost := p.CDNHost
	if cdnHost == "" {
		cdnHost = p.SigningHost
	}

	cdnPath := "/" + strings.TrimPrefix(p.Key, "/")
	return fmt.Sprintf("https://%s%s?%s", cdnHost, cdnPath, finalQuery)
}

func buildCanonicalQueryString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(params))
	for _, k := range keys {
		parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(params[k]))
	}
	return strings.Join(parts, "&")
}

func sha256hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func deriveSigningKey(secret, date, region string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(date))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte("s3"))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}
