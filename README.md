[![Docker](https://github.com/Ogglord/comp2unraid/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/Ogglord/comp2unraid/actions/workflows/docker-publish.yml) [![Go](https://github.com/Ogglord/comp2unraid/actions/workflows/go.yml/badge.svg)](https://github.com/Ogglord/comp2unraid/actions/workflows/go.yml)

# comp2unraid
Convert docker compose templates to unraid template


## Usage

 - Have your docker-compose.yml ready, as an URL _https://...docker-compose.yml_ or a local file on your server
 - SSH to your Unraid server or open a shell terminal in your browser
 - Execute it without installing using docker run, download the binary or compile from source, whatever you prefer.


### Run using docker

Preview the xml first, using the -n (dry run) flag
```bash
docker run ghcr.io/ogglord/comp2unraid -n "https://raw.githubusercontent.com/user/r/docker-compose.yml"
```
Then; pipe the output to an xml file
```bash
docker run ghcr.io/ogglord/comp2unraid -n "https://raw.githubusercontent.com/user/r/docker-compose.yml" > my-template.xml
```
Note: if your docker-compose contains multiple services, you will have to split the xml manually, or run the binary, see below, which generates one xml file per service


### Download binary

 - Download the latest release from Github
 - ```./comp2unraid "https://raw.githubusercontent.com/user/r/docker-compose.yml"```

### Compile from source

 - Install go v1.23.1
 - Clone this repo
 - ```go build```
 - ```./comp2unraid "https://raw.githubusercontent.com/user/r/docker-compose.yml"```



