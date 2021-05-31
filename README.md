# devnet

Spin up, deliver a payload, issue commands to, and collect the logs of all of those droplets using a config.

## Install
currently using pulumi to manage spinning up and down droplets. We could just use the digital ocean go API, but this seemed like a better way to make sure that we don't spin up too many droplets. This can change.
install `pulumi` [here](https://www.pulumi.com/docs/get-started/install/)

clone or fork this repo

## Export your DO access token

```
export DIGITALOCEAN_ACCESS_TOKEN="your token"
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
here's an example config
```json
{
    "ssh_key_id": "put your DO ssh key finger print here", // upload your ssh public key to your DO account
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
## Usage

go to the pulumi directory

```
cd ./pulumi
```




after setting up pulumi, spin up the nodes by following the prompts

```
pulumi up
```

compile `devnet` by calling `go build` in this directory

call `devnet init` to deliver the specified payloads to the droplets, call the initial commands, and then it will start saving the logs of those commands to the files specified in the config. This should overwrite any preexisting payloads, so no need to spin up the nodes everytime.

don't forget to change back the pulumi directory and spin down the nodes by following the prompts

```
pulumi destroy
```