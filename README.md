# RestProxy
Proxy server for emulation delays and error status codes

```
Usage:
  -backend_url string
        Backend URL (default "127.0.0.1:8080")
  -block_config string
        Serialized dict for blocking backend endpoints where keys are patterns of endpoints, values are response status codes,for instance: "{\"profile\": 404}" (default "{}")
  -delay_config string
        Serialized dict for delaying backend endpoints where keys are patterns of endpoints, values are delays in seconds,for instance: "{\"profile\": 35}" (default "{}")
  -localPort int
        Local port (default 5050)
```
