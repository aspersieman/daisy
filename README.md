# Dev

Install go
Install air
Get .env: cp .env.local .env
make dev

# Building and running your application

When you're ready, start your application by running:
`docker compose up --build`.

Your application will be available at http://localhost:8069.

# Deploying your application to the cloud

First, build your image, e.g.: `docker build -t ghcr.io/aspersieman/daisy:latest .`.
If your cloud uses a different CPU architecture than your development
machine (e.g., you are on a Mac M1 and your cloud provider is amd64),
you'll want to build the image for that platform, e.g.:
`docker build --platform=linux/amd64 -t ghcr.io/aspersieman/daisy:latest .`.

Then, push it to your registry, e.g. `docker push ghcr.io/aspersieman/daisy:latest`.

Consult Docker's [getting started](https://docs.docker.com/go/get-started-sharing/)
docs for more detail on building and pushing.

# References
* [Docker's Go guide](https://docs.docker.com/language/golang/)

# Swarm

## Server

- Remove extant dockers: `for pkg in docker.io docker-doc docker-compose podman-docker containerd runc; do sudo apt-get remove $pkg; done`
- Add docker repos: ```sudo apt-get update
sudo apt-get install ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update```
- Install docker: `sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin`
- Add relevant user to docker group: `sudo usermod -aG docker $USER`
- Init a docker swarm: `docker swarm init`

- Setup application's configuration storage:
    - `mkdir -p /home/deployment/config/daisy/`
    - Client: `scp ~/.docker/config.json deployment@vu-za-app-1:/home/deployment/config/`
    - Client: `scp .env deployment@vu-za-app-1:/home/deployment/config/daisy/`

## Client
- Create docker context for new swarm: `docker context create vu-za-app-1 --docker "host=ssh://deployment@139.84.239.128"`
NOTE: Ensure you've ssh access to the host and have ssh'd in before
- Use new context `docker context use vu-za-app-1`
- Deploy to new swarm: `docker stack deploy -c compose.prod.yaml daisy`
- Normal compose actions apply: `docker compose -f compose.prod.yaml up -d`, `docker compose -f compose.prod.yaml stop`
- Use back to default context `docker context use default`
- Remove a swarm: `docker service scale daisy_daisy=0 && docker service rm daisy_daisy"

## Github actions

- (Create a personal access token)[https://docs.github.com/en/enterprise-server@2.22/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token]

## ufw

- Deny all inbound requests: `sudo ufw default deny incoming`
- Allow all outgoing requests `sudo ufw default allow outgoing`
- Allow connections to OpenSSH `sudo ufw allow OpenSSH`
- Allow HTTP connections (port 80): `sudo ufw allow 80`
 - Enable ufw `sudo ufw enable`

## Docker manual

 - Create github personal access token (classic): https://github.com/settings/tokens
 - Login to registry: `docker login --username aspersieman@gmail.com --password <my password> ghcr.io`
 - Build, tag container `docker build -t daisy . -t ghcr.io/aspersieman/daisy:latest`
 - Push to registry: `docker push ghcr.io/aspersieman/daisy:latest`

## Install
 - fail2ban
 - rkhunter

# References
 - https://www.design.com/maker/logo
