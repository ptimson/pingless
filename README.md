<p align="center">
  <img src="assets/logo.png" alt="Pingless Logo" width="200" />
</p>

<h1 align="center">Pingless</h1>

<p align="center">
  <b>The scriptable ping watch bot</b><br/>
  Reboot routers, restart services, or run any script when pings fail.
</p>

## Why?

I built **Pingless** because my Zyxel 5G router would occasionally stop responding to the internet.  
The only way to fix it was to reboot the router. Instead of doing this manually, Pingless pings a host every minute, and
if it fails repeatedly, it triggers a script to fix the problem.

While my use case is rebooting a Zyxel router, you can make it run **any script you want**.  
For example, it could restart a container, reboot a server, or send you a notification.

## How it works

- Pingless pings a host (default `8.8.8.8`) at regular intervals.
- If it fails **N times in a row**, it executes a script you provide.
- The script can do anything:  
  Example → Zyxel reboot script: [`scripts/zyxel-reboot-router.sh`](./scripts/zyxel-reboot-router.sh)
- After running the script, pingless waits `CMD_WAIT_INTERVAL` before resuming pings. Useful for giving a router or service time to restart.

## Quick Start

### Docker Compose

```yml
services:
  pingless:
    image: ptimson/pingless:latest
    container_name: pingless
    cap_add:
      - NET_RAW # required to send ICMP ping
    environment:
      # Target to ping
      PING_HOST: 8.8.8.8

      # How often to run the main check loop
      PING_INTERVAL: 60s

      # Timeout for each ping
      PING_TIMEOUT: 3s

      # Retry delay between consecutive failed attempts
      RETRY_DELAY: 5s

      # Number of consecutive failures before triggering script
      MAX_FAILURES: 10

      # How long to wait before testing ping once script has executed
      # e.g. time for router to fully restart 
      CMD_WAIT_INTERVAL: 2m

    volumes:
      # Script that is executed on ping failure
      # Only add :ro if file is already executable (chmod +x)
      - ./on-ping-fail.sh:/app/on-ping-fail.sh:ro

```

### Docker Run

```bash
docker run --rm -it \
  --cap-add=NET_RAW \                      # required for ICMP ping  
  --env PING_HOST=8.8.8.8 \                # target to ping  
  --env PING_INTERVAL=60s \                # how often to ping  
  --env RETRY_DELAY=5s \                   # delay between failed retries  
  --env MAX_FAILURES=10 \                  # how many fails before trigger
  --env CMD_WAIT_INTERVAL=2m \             # wait time once script has executed  
  --volume ./on-ping-fail.sh:/app/on-ping-fail.sh \
  ptimson/pingless:latest  
```

## Configuration

| Env Var             | Default   | Description                                                |
|---------------------|-----------|------------------------------------------------------------|
| `PING_HOST`         | `8.8.8.8` | Host to ping                                               |
| `PING_INTERVAL`     | `60s`     | Interval between pings                                     |
| `PING_TIMEOUT`      | `3s`      | Timeout per ping attempt                                   |
| `RETRY_DELAY`       | `5s`      | Delay between failed retries                               |
| `MAX_FAILURES`      | `10`      | Consecutive failures before running script                 |
| `CMD_WAIT_INTERVAL` | `2m`      | Wait time to ping once script has run (e.g. router reboot) |

> On failure, Pingless will run `/app/on-ping-fail.sh`.  
> Mount your script into that path and make sure it’s executable (`chmod +x`).

## Example: Zyxel Router Reboot

See example of how I reboot my router in `scripts/zyxel-reboot-router.sh`

## License

MIT