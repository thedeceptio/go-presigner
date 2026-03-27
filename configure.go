package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func runConfigure(profile string) error {
	existing, err := LoadConfig(profile)
	if err != nil {
		return err
	}

	fmt.Printf("Configuring profile [%s]\n", profile)
	fmt.Println("Press Enter to keep the existing value shown in brackets.")
	fmt.Println()

	cfg := &Config{}

	cfg.AccessKeyID = prompt("AWS Access Key ID", maskValue(existing.AccessKeyID, false), existing.AccessKeyID)
	cfg.SecretAccessKey = prompt("AWS Secret Access Key", maskValue(existing.SecretAccessKey, true), existing.SecretAccessKey)
	cfg.Region = promptDefault("Region", existing.Region, "us-east-1")
	cfg.Bucket = promptDefault("Bucket", existing.Bucket, "")
	cfg.SigningHost = promptDefault("Signing Host", existing.SigningHost, "s3.amazonaws.com")
	cfg.CDNHost = promptDefault("CDN Host (leave blank to use Signing Host)", existing.CDNHost, "")

	expiresDefault := 3600
	if existing.ExpiresIn > 0 {
		expiresDefault = existing.ExpiresIn
	}
	expiresStr := promptDefault("Expires In (seconds)", strconv.Itoa(expiresDefault), strconv.Itoa(expiresDefault))
	cfg.ExpiresIn, err = strconv.Atoi(expiresStr)
	if err != nil || cfg.ExpiresIn <= 0 {
		return fmt.Errorf("expires_in must be a positive number, got %q", expiresStr)
	}

	if err := SaveConfig(profile, cfg); err != nil {
		return err
	}

	path, _ := configPath()
	fmt.Printf("\nConfiguration saved to %s\n", path)
	return nil
}

// prompt shows a prompt with the display hint and returns user input or fallback.
func prompt(label, display, fallback string) string {
	if display != "" {
		fmt.Printf("%s [%s]: ", label, display)
	} else {
		fmt.Printf("%s: ", label)
	}

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return fallback
	}
	return input
}

// promptDefault shows a prompt; if existing is non-empty it's used as display+fallback,
// otherwise defaultVal is the fallback.
func promptDefault(label, existing, defaultVal string) string {
	display := existing
	fallback := existing
	if existing == "" {
		display = defaultVal
		fallback = defaultVal
	}
	return prompt(label, display, fallback)
}

// maskValue masks a credential value for display.
func maskValue(v string, mask bool) string {
	if v == "" {
		return ""
	}
	if !mask {
		// Show last 4 chars of access key ID.
		if len(v) > 4 {
			return "****************" + v[len(v)-4:]
		}
		return "****************"
	}
	// Always fully mask secret.
	return "****************"
}
