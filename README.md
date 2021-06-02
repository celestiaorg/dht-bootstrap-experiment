# DHT Bootstrap Experiment

This is a template repo to spin up digital ocean droplets, deliver payloads, issue commands to, and collect the logs in real time of those droplets using a config. Currently, it is configured to run an experiment to collect rough data on how long it takes to randomly sample from block data produced by a single lazyledger-core node.

## Install
### Install and setup pulumi
https://www.pulumi.com/docs/get-started/install/

clone or fork this repo

### Upload a SSH public key to digital ocean

get the fingerprint of the ssh key uploaded to DO. [instructions](https://docs.digitalocean.com/products/droplets/how-to/add-ssh-keys/to-account/)

## Usage

### Configure your deployment

### Configuration

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

edit `config.json` to add in your DO fingerprint, and path to output
```json
{
    "ssh_key_id": "***your DO ssh key fingerprint here***",
    "tag": "DHT-bootstrap",
    "droplets": {
        "validator1nyc3": {
            "location": "nyc3",
            "size": "s-1vcpu-1gb",
            "name": "val1",
            "type": 0,
            "payload": "../payloads/validator",
            "init_commands": [                
                "/root/validator/tendermint init",
                "source /root/validator/init.sh",
                "sed -i 's_tcp://127.0.0.1:26657_tcp://0.0.0.0:26657_g' /root/.tendermint/config/config.toml",
                "./validator/tendermint node --proxy-app=kvstore"
            ], 
            "output": "../logs/validator1-nyc3.log"
        },
        "light1tor1": {
            "location": "tor1",
            "size": "s-1vcpu-1gb",
            "name": "light1",
            "type": 2,
            "payload": "../payloads/light",
            "init_commands": [
                "source /root/light/init.sh"
            ],
            "output": "../logs/light1-tor1.log"
        },
        "dht1sgp1": {
            "location": "sgp1",
            "size": "s-1vcpu-1gb",
            "name": "light1",
            "type": 3,
            "payload": "../payloads/dht",
            "init_commands": [
                "source /root/dht/init.sh"
            ],
            "output": "../logs/dht1-sgp1.log"
        }
        
    }
}

```

## Export your DO access token

```sh
export DIGITALOCEAN_ACCESS_TOKEN="your token"
# the program will ask you to type your ssh password during execution if you don't want to export it. 
# Can also set to "nil" to ignore prompt and use "" as a password. 
# It also looks for the signing key in the default location $HOME/.ssh/ and this has to be changed manually as of now.
export SSH_PASS="your ssh pass"
```

go to the pulumi directory

```sh
cd pulumi
```

after setting up pulumi, spin up the nodes by following the prompts

```sh
pulumi up
```

compile by calling `go build -o arbitrary-binary-name` in the root of this directory

call 
```sh
aritrary-binary-name init config.json
``` 

to deliver the specified payloads to the droplets (including `public_ipv4s.json` and `public_ipv4s.sh` files with all the deployed droplet's public IPs), call the init command, and then it will start saving the logs of those commands to the files specified in the config. This should overwrite any preexisting payloads, so no need to spin up and destroy droplets everytime.

don't forget to change back the pulumi directory and spin down the nodes by following the prompts after calling 

```sh
pulumi destroy
```