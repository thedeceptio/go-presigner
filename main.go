package main

import (
	"fmt"
	"os"
)

const usage = `go-presigner — AWS S3 presigned URL generator

Usage:
  go-presigner configure                    interactive setup
  go-presigner configure list               show current config
  go-presigner configure set <field> <val>  set a single config field
  go-presigner presign <key> [flags]        generate a presigned URL

Global flags:
  --profile <name>   config profile to use (default: "default")

Fields for 'configure set':
  aws_access_key_id, aws_secret_access_key, region, bucket,
  signing_host, cdn_host, expires_in

Examples:
  go-presigner configure
  go-presigner configure list
  go-presigner configure set bucket my-bucket
  go-presigner presign path/to/file.pdf
  go-presigner presign path/to/file.pdf --expires-in 300
  go-presigner --profile staging presign path/to/file.pdf
`

func main() {
	args := os.Args[1:]

	// Extract --profile <name> from anywhere in the args before routing.
	profile := "default"
	filtered := args[:0]
	for i := 0; i < len(args); i++ {
		if args[i] == "--profile" || args[i] == "-profile" {
			if i+1 >= len(args) {
				die("--profile requires a value")
			}
			profile = args[i+1]
			i++
		} else {
			filtered = append(filtered, args[i])
		}
	}
	args = filtered

	if len(args) == 0 {
		fmt.Print(usage)
		os.Exit(0)
	}

	var err error

	switch args[0] {
	case "configure":
		err = handleConfigure(profile, args[1:])

	case "presign":
		if len(args) < 2 {
			die("presign requires an object key\nUsage: go-presigner presign <key> [flags]")
		}
		err = runPresign(profile, args[1], args[2:])

	case "help", "--help", "-h":
		fmt.Print(usage)

	default:
		die(fmt.Sprintf("unknown command %q\n\n%s", args[0], usage))
	}

	if err != nil {
		die(err.Error())
	}
}

func handleConfigure(profile string, args []string) error {
	if len(args) == 0 {
		return runConfigure(profile)
	}

	switch args[0] {
	case "list":
		return PrintConfig(profile)

	case "set":
		if len(args) < 3 {
			return fmt.Errorf("usage: go-presigner configure set <field> <value>")
		}
		return SetConfigField(profile, args[1], args[2])

	default:
		return fmt.Errorf("unknown configure subcommand %q\nUsage: go-presigner configure [list|set]", args[0])
	}
}

func die(msg string) {
	fmt.Fprintln(os.Stderr, "error:", msg)
	os.Exit(1)
}
