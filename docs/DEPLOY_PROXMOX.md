# Deploy Primordia On Proxmox

This guide covers two deployment styles:

- Docker inside a VM on Proxmox (recommended)
- Docker inside an LXC on Proxmox (works, but needs extra host settings)

## Option 1: Docker Host VM (Recommended)

Use a Debian/Ubuntu VM as your Docker host.

### 1. Prepare the VM

- Install Docker Engine and Docker Compose plugin.
- Ensure ports are allowed in VM firewall and Proxmox firewall.

### 2. Publish the image

From your Primordia source machine:

```bash
make docker-deploy IMAGE_NAME=ghcr.io/<owner>/primordia IMAGE_TAG=latest
```

### 3. Run on the Docker VM

On the VM:

```bash
docker login ghcr.io
docker pull ghcr.io/<owner>/primordia:latest
docker run -d --name primordia --restart unless-stopped -p 8080:8080 ghcr.io/<owner>/primordia:latest
```

Health check:

```bash
docker logs --tail 100 primordia
curl -i http://127.0.0.1:8080/ws
```

Expected behavior:

- Process starts and logs: Primordia engine running on :8080
- /ws upgrades to websocket (curl will not fully establish ws, but route should exist)

## Option 2: Docker In LXC (Advanced)

This can run well, but requires nested-container support.

### 1. Create LXC

- Debian 12 template (or Ubuntu 24.04)
- Enable network bridge and static or DHCP IP

### 2. Enable required LXC features on Proxmox

In Proxmox for the container:

- Options -> Features: enable `Nesting`
- Options -> Features: enable `Keyctl`

Or in config `/etc/pve/lxc/<CTID>.conf`:

```conf
features: nesting=1,keyctl=1
```

For some setups you may also need:

```conf
lxc.apparmor.profile: unconfined
lxc.cgroup2.devices.allow: a
lxc.cap.drop:
```

Then restart the LXC.

### 3. Install Docker inside LXC

Inside the LXC:

```bash
apt update
apt install -y ca-certificates curl gnupg
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  $(. /etc/os-release && echo $VERSION_CODENAME) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt update
apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

### 4. Run Primordia inside LXC Docker

```bash
docker login ghcr.io
docker pull ghcr.io/<owner>/primordia:latest
docker run -d --name primordia --restart unless-stopped -p 8080:8080 ghcr.io/<owner>/primordia:latest
```

## Compose-Based Deployment

You can also deploy with compose from this repo:

```bash
IMAGE=ghcr.io/<owner>/primordia:latest HOST_PORT=8080 docker compose up -d
```

Using Make targets:

```bash
IMAGE=ghcr.io/<owner>/primordia:latest make compose-up
```

## Reverse Proxy (Optional)

If exposed publicly, front it with Traefik, Caddy, or Nginx and proxy websocket upgrades to port 8080.

## Recommendation

For a Proxmox datacenter, run Primordia on a dedicated Docker VM first for lowest operational risk.
If you want maximum density and are comfortable with nested-container troubleshooting, LXC + Docker is viable.