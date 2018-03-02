# Telegraf Plugin: nginx_upstream_check

Nginx Upstream Check is a input's module that support the nginx_http_upstream_check_module. The use this plugin you will need a compile with nginx_http_upstream_check_module. For more information about the nginx_http_upstream_check_module. [click here](https://github.com/yaoweibin/nginx_upstream_check_module)

### Configuration:

```
# Read Nginx upstream health check status information
[[inputs.nginx_upstream_check]]
  # An array of Nginx status of upstream URI to gather stats.
  urls = ["http://localhost/status?format=json"]

  # HTTP response timeout (default: 5s)
  response_timeout = "5s"
```

### Measurements & Fields:

- nginx_upstream_check
  - generation
  - host
  - port
  - server
  - total
  
- nginx_upstream_check_peer
  - index
  - upstream_name
  - upstream_server
  - status
  - rise
  - fall
  - type
  - port
  - host

### Tags:

- nginx_upstream_check
  - host
  - port
  - server

- nginx_upstream_check_peer
  - host
  - index
  - upstream_name
  - upstream_server

### Example Output:

Using this configuration:
```
# Read Nginx upstream health check status information
[[inputs.nginx_upstream_check]]
  # An array of Nginx status of upstream URI to gather stats.
  urls = ["http://localhost/status?format=json"]

  # HTTP response timeout (default: 5s)
  response_timeout = "5s"
```

When run with:
```
./telegraf -config telegraf.conf -input-filter nginx_upstream_check -test
```

It produces:
```
* Plugin: inputs.nginx_upstream_check, Collection 1
> nginx_upstream_check,server=192.168.96.45,port=8080,host=Zeus total=2i,generation=1i 1519995939000000000
> nginx_upstream_check_peer,host=Zeus,upstream_name=upstreamcluster,upstream_server=192.168.96.20:8180,index=0 type="http",port=0i,status="down",rise=0i,fall=1477i 1519995939000000000
> nginx_upstream_check_peer,upstream_name=upstreamcluster,upstream_server=192.168.96.20:8280,index=1,host=Zeus fall=1477i,type="http",port=0i,status="down",rise=0i 1519995939000000000
```

### Reference material

Subsequent versions of status response structure available here:

- [nginx_upstream_check_module](https://github.com/yaoweibin/nginx_upstream_check_module)