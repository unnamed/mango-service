# mango-service

A simple file storage service, files in the mango-service are temporal, limited and 
can be downloaded only once. There is also a limit of a single file per IP address.

### Usage
To build `mango-service`, just use `./go build`. Then use the built executable:
```bash
./mango-service --port=2069 --dir=data
```
The executable accepts multiple flags

| Option       | Default Value          | Description                                      |
|--------------|------------------------|--------------------------------------------------|
| `port`       | `2069`                 | The HTTP server port                             |
| `dir`        | `data`                 | The directory that will contain the stored files |
| `size-limit` | `5000000` *(5 MB)*     | The file size limit (in bytes)                   |
| `lifetime`   | `300000` *(5 minutes)* | The file lifetime (in milliseconds)              |