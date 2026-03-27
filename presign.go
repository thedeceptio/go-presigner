package main

import (
	"flag"
	"fmt"
	"os"
)

func runPresign(profile, key string, args []string) error {
	fs := flag.NewFlagSet("presign", flag.ContinueOnError)

	var (
		region      = fs.String("region", "", "AWS region (overrides config)")
		bucket      = fs.String("bucket", "", "S3 bucket name (overrides config)")
		signingHost = fs.String("signing-host", "", "Signing host (overrides config)")
		cdnHost     = fs.String("cdn-host", "", "CDN host (overrides config)")
		expiresIn   = fs.Int("expires-in", 0, "URL expiry in seconds (overrides config)")
	)

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: go-presigner presign <key> [flags]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Flags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := LoadConfig(profile)
	if err != nil {
		return err
	}

	// Merge: flag overrides config, config fills in blanks.
	params := PresignParams{Key: key}

	params.AccessKeyID = coalesce(os.Getenv("AWS_ACCESS_KEY_ID"), cfg.AccessKeyID)
	params.SecretAccessKey = coalesce(os.Getenv("AWS_SECRET_ACCESS_KEY"), cfg.SecretAccessKey)

	params.Region = firstNonEmpty(*region, cfg.Region, "us-east-1")
	params.Bucket = firstNonEmpty(*bucket, cfg.Bucket)
	params.SigningHost = firstNonEmpty(*signingHost, cfg.SigningHost, "s3.amazonaws.com")
	params.CDNHost = firstNonEmpty(*cdnHost, cfg.CDNHost)

	params.ExpiresIn = *expiresIn
	if params.ExpiresIn == 0 {
		params.ExpiresIn = cfg.ExpiresIn
	}
	if params.ExpiresIn == 0 {
		params.ExpiresIn = 3600
	}

	if params.AccessKeyID == "" {
		return fmt.Errorf("missing aws_access_key_id — run 'go-presigner configure' or set AWS_ACCESS_KEY_ID")
	}
	if params.SecretAccessKey == "" {
		return fmt.Errorf("missing aws_secret_access_key — run 'go-presigner configure' or set AWS_SECRET_ACCESS_KEY")
	}
	if params.Bucket == "" {
		return fmt.Errorf("missing bucket — run 'go-presigner configure' or pass --bucket")
	}

	url := CreatePresignedURL(params)
	fmt.Println(url)
	return nil
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
