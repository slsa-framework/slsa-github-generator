# setup-token

The token has the following format:

```text
token := B64_BUNDLE.B64_TOKEN
```

where

```text
B64_TOKEN := base64(JSON_RAW_TOKEN)
B64_BUNDLE := base64(Sign(B64_TOKEN))
```
