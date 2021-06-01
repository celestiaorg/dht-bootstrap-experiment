# devnet

This is a quick repo to spin up digital ocean droplets, deliver payloads, issue commands to, and collect the logs in real time of those droplets using a config.

## Install
currently using pulumi to manage spinning up and down droplets. We could just use the digital ocean go API or terraform, but this seemed like a better way to make sure that we don't spin up too many droplets, while not using the terraform DSL. This can change.

install `pulumi` [here](https://www.pulumi.com/docs/get-started/install/)

clone or fork this repo

## Upload a SSH public key to digital

have a public ssh key uploaded to DO [uploaded to DO](https://docs.digitalocean.com/products/droplets/how-to/add-ssh-keys/to-account/)

## Export your DO access token

```sh
export DIGITALOCEAN_ACCESS_TOKEN="your token"
# the program will ask you to type your ssh password during execution if you don't want to export it. 
# Can also set to "nil" to ignore prompt and use "" as a password. 
# It also looks for the signing key in the default location $HOME/.ssh/ and this has to be changed manually as of now.
export SSH_PASS="your ssh pass"
```

## Configuration

The config defines everything needed to spin up. Save it to this repo as `config.json`

```go
// Config structures the data used to configure a deployment
type Config struct {
	Droplets map[string]Droplet `json:"droplets"`
	// SSHKeyID is the fingerprint of the ssh key preloaded into digital ocean
	SSHKeyID string `json:"ssh_key_id"`
	// Tag is used to idendify droplets that belong to this deployment
	Tag string `json:"tag"`
}

// Droplet specifies a single droplet
type Droplet struct {
	// Location is the digital ocean region for the droplet
	// ie "nyc3"
	Location string `json:"location"`
	// Size indicates how big the droplet size should be
	// ie "s-1vcpu-1gb"
	Size string `json:"size"`
	// Type: Validator || Full || LightClient || DHT
	Type NodeType `json:"droplet_type"` // type can probably be removed
	// Payload is the path to directory that is to be copied to the server
	Payload string `json:"payload"`
	// InitCommands are the ssh commands run after payload is delivered
	InitCommands []string `json:"init_commands"`
	// Output is the path to output file
	Output string `json:"output"`
	Drop   godo.Droplet
}
```

## Usage

### Configure your deployment
here's an example config
```json
{
    "ssh_key_id": "put your DO ssh key finger print here",
    "tag": "devnet",
    "droplets": {
        "validator1": {
            "location": "nyc3",
            "size": "s-1vcpu-1gb",
            "name": "val1",
            "type": 0,
            "payload": "path/to/payloads/validator/",
            "init_commands": [
                "source ./validator/init.sh"
            ], 
            "output": "/path/to/live/log/output/validator1.txt"
        },
        "validator2": {
            "location": "nyc3",
            "size": "s-1vcpu-1gb",
            "name": "val2",
            "type": 0,
            "payload": "path/to/payloads/validator/",
            "init_commands": [
                "source ./validator/init.sh",
                "echo 'Im validator 2'",
            ], 
            "output": "/path/to/live/log/output/validator2.txt"
        }
        
    }
}

```

go to the pulumi directory

```sh
cd ./pulumi
```

after setting up pulumi, spin up the nodes by following the prompts

```sh
pulumi up
```

compile `devnet` by calling `go build` in this directory

call 
```sh
devnet init /path/to/config.json
``` 

to deliver the specified payloads to the droplets (including a `public_ipv4s.json` file with all the deployed droplet's public IPs), call the init command, and then it will start saving the logs of those commands to the files specified in the config. This should overwrite any preexisting payloads, so no need to spin up and destroy droplets everytime.

don't forget to change back the pulumi directory and spin down the nodes by following the prompts after calling 

```sh
pulumi destroy
```