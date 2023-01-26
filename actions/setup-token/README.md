# setup-token

The token has the following format:
```
token := B64_BUNDLE.B64_TOKEN
```

where
```
B64_TOKEN := base64(JSON_RAW_TOKEN)
B64_BUNDLE := base64(Sign(B64_TOKEN))
```