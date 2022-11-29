# GitHub Actions Metrics

## Creating GitHub HMAC Signature

```bash
cat data.json | openssl sha256 -hmac "my-secret-string"

# Output:
db9b9ea04baa347372ba77cdc9877474efe276347969a9b4790754d28bfe69e4
```

Use this value in the `X-Hub-Signature-256` request header as follows:

```bash
X-Hub-Signature-256: sha256=db9b9ea04baa347372ba77cdc9877474efe276347969a9b4790754d28bfe69e4
```

### Example Request

```bash
curl \
 -H "X-Github-Delivery: $(uuidgen)" \
 -H "X-Github-Event: workflow-run" \
 -H "X-Hub-Signature-256: sha256=$(cat data.json | openssl sha256 -hmac "my-secret-string")" \
 -d "@data.json" \
 http://localhost:8080
```

URL
JSON encoded payload
Create/paste secret (should also be in secret manager)
