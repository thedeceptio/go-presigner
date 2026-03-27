# go-presigner

AWS S3 Signature V4 presigned URL generator — configure once, presign anywhere.

Works with AWS S3 and S3-compatible storage (MinIO, DigitalOcean Spaces, etc.). Supports a separate CDN host in the returned URL while signing against the storage host.

## Install

**With Go:**
```bash
go install github.com/thedeceptio/go-presigner@latest
```

**Pre-built binaries:** Download from [Releases](https://github.com/thedeceptio/go-presigner/releases).

## Usage

### Configure

```bash
go-presigner configure
```

```
Configuring profile [default]
Press Enter to keep the existing value shown in brackets.

AWS Access Key ID: AKIAIOSFODNN7EXAMPLE
AWS Secret Access Key: ****************
Region [us-east-1]:
Bucket: my-bucket
Signing Host [s3.amazonaws.com]: s3.us-east-1.amazonaws.com
CDN Host (leave blank to use Signing Host): cdn.example.com
Expires In (seconds) [3600]:

Configuration saved to /home/user/.go-presigner/config
```

### Presign

```bash
go-presigner presign "path/to/file.pdf"
# https://cdn.example.com/path/to/file.pdf?X-Amz-Algorithm=...
```

Override any config value on the fly:
```bash
go-presigner presign "path/to/file.pdf" --expires-in 300
go-presigner presign "path/to/file.pdf" --cdn-host "other.cdn.com" --bucket "other-bucket"
```

### Multiple profiles

```bash
go-presigner --profile staging configure
go-presigner --profile staging presign "path/to/file.pdf"
```

### Other commands

```bash
go-presigner configure list                        # show config (secret masked)
go-presigner configure set bucket my-bucket        # set a single field
go-presigner configure set expires_in 7200
```

## Config fields

| Field | Description | Default |
|-------|-------------|---------|
| `aws_access_key_id` | AWS access key ID | — |
| `aws_secret_access_key` | AWS secret access key | — |
| `region` | AWS region | `us-east-1` |
| `bucket` | S3 bucket name | — |
| `signing_host` | Host used for signature computation | `s3.amazonaws.com` |
| `cdn_host` | Host used in the returned URL | same as `signing_host` |
| `expires_in` | URL expiry in seconds | `3600` |

Credentials can also be provided via `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables, which take priority over the config file.

Config is stored at `~/.go-presigner/config` in INI format with `0600` permissions.

## Signing host vs CDN host

`signing_host` is the hostname the storage service uses to verify the signature (e.g. `s3.us-east-1.amazonaws.com` or `minio.internal`).

`cdn_host` is the hostname in the URL you hand to clients (e.g. `cdn.example.com`). This lets you serve presigned URLs through CloudFront or a reverse proxy without breaking the signature.

## License

MIT
