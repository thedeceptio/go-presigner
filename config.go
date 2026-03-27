package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Bucket          string
	SigningHost     string
	CDNHost         string
	ExpiresIn       int
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".go-presigner"), nil
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config"), nil
}

// LoadConfig reads the config file and returns the config for the given profile.
func LoadConfig(profile string) (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	defer f.Close()

	sections := parseINI(f)

	section, ok := sections[profile]
	if !ok {
		return &Config{}, nil
	}

	cfg := &Config{
		AccessKeyID:     section["aws_access_key_id"],
		SecretAccessKey: section["aws_secret_access_key"],
		Region:          section["region"],
		Bucket:          section["bucket"],
		SigningHost:     section["signing_host"],
		CDNHost:         section["cdn_host"],
	}

	if v, ok := section["expires_in"]; ok {
		cfg.ExpiresIn, _ = strconv.Atoi(v)
	}

	return cfg, nil
}

// SaveConfig writes the config for the given profile to the config file.
func SaveConfig(profile string, cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path, err := configPath()
	if err != nil {
		return err
	}

	// Read existing sections so we don't overwrite other profiles.
	sections := map[string]map[string]string{}
	if f, err := os.Open(path); err == nil {
		sections = parseINI(f)
		f.Close()
	}

	sections[profile] = map[string]string{
		"aws_access_key_id":     cfg.AccessKeyID,
		"aws_secret_access_key": cfg.SecretAccessKey,
		"region":                cfg.Region,
		"bucket":                cfg.Bucket,
		"signing_host":          cfg.SigningHost,
		"cdn_host":              cfg.CDNHost,
		"expires_in":            strconv.Itoa(cfg.ExpiresIn),
	}

	return writeINI(path, sections)
}

// SetConfigField sets a single field in the config for a profile.
func SetConfigField(profile, field, value string) error {
	cfg, err := LoadConfig(profile)
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = &Config{}
	}

	switch field {
	case "aws_access_key_id":
		cfg.AccessKeyID = value
	case "aws_secret_access_key":
		cfg.SecretAccessKey = value
	case "region":
		cfg.Region = value
	case "bucket":
		cfg.Bucket = value
	case "signing_host":
		cfg.SigningHost = value
	case "cdn_host":
		cfg.CDNHost = value
	case "expires_in":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("expires_in must be a number")
		}
		cfg.ExpiresIn = n
	default:
		return fmt.Errorf("unknown config field %q\nvalid fields: aws_access_key_id, aws_secret_access_key, region, bucket, signing_host, cdn_host, expires_in", field)
	}

	return SaveConfig(profile, cfg)
}

// PrintConfig prints the config for a profile, masking the secret key.
func PrintConfig(profile string) error {
	cfg, err := LoadConfig(profile)
	if err != nil {
		return err
	}

	secret := cfg.SecretAccessKey
	if len(secret) > 4 {
		secret = "****************" + secret[len(secret)-4:]
	} else if secret != "" {
		secret = "****************"
	}

	fmt.Printf("[%s]\n", profile)
	fmt.Printf("aws_access_key_id     = %s\n", cfg.AccessKeyID)
	fmt.Printf("aws_secret_access_key = %s\n", secret)
	fmt.Printf("region                = %s\n", cfg.Region)
	fmt.Printf("bucket                = %s\n", cfg.Bucket)
	fmt.Printf("signing_host          = %s\n", cfg.SigningHost)
	fmt.Printf("cdn_host              = %s\n", cfg.CDNHost)
	fmt.Printf("expires_in            = %d\n", cfg.ExpiresIn)
	return nil
}

// parseINI parses a simple INI file into a map of section -> key -> value.
func parseINI(f *os.File) map[string]map[string]string {
	sections := map[string]map[string]string{}
	current := ""

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			current = line[1 : len(line)-1]
			if sections[current] == nil {
				sections[current] = map[string]string{}
			}
			continue
		}
		if current == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		sections[current][k] = v
	}
	return sections
}

// writeINI writes all sections to the config file.
func writeINI(path string, sections map[string]map[string]string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write "default" first if present, then remaining profiles alphabetically.
	order := []string{}
	if _, ok := sections["default"]; ok {
		order = append(order, "default")
	}
	for k := range sections {
		if k != "default" {
			order = append(order, k)
		}
	}

	w := bufio.NewWriter(f)
	for i, name := range order {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "[%s]\n", name)
		fields := []string{
			"aws_access_key_id", "aws_secret_access_key",
			"region", "bucket", "signing_host", "cdn_host", "expires_in",
		}
		for _, k := range fields {
			if v, ok := sections[name][k]; ok {
				fmt.Fprintf(w, "%-22s= %s\n", k, v)
			}
		}
	}
	return w.Flush()
}
