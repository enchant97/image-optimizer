# Image Optimizer
A service allowing for the optimizations of images, designed to integrate as a backend micro-service. Supporting new formats and clustering.

> Still a WIP, may not be suited for production


## Features
- Configurable optimization settings (using YAML)
    - Resize images
    - Convert
        - JPEG
        - WEBP
        - AVIF
    - Quality setting
- Upload size limiting
- Cluster support (many producers and/or many consumers)
- Scan existing folder of originals
- Access control for API
- REST API
    - Trigger job to load from storage
    - Trigger job directly from image upload
- Job queue powered via RabbitMQ


## Future Features
- S3 support
- Zones (fancy folders, with access control)


## Structure
There is a specific structure to the optimised output, shown below:

```
originals/
    /example.jpg

optimized/
    /<path>/<original name>@<optimization name>.<optimization format>
    /example.jpg@large.webp
```

Running the app is quite flexible. With the ability to run many producers and many consumers, whilst allowing a simple setup of a combined producer/consumer service. This allows the service to scale out.

Examples:

- 3 producer nodes and 8 consumer nodes.
- 1 producer node and 1 consumer node
- 1 combined node

Consumers require no extra configuration to handle scaling, since this is handled via RabbitMQ queues. However running a publisher requires something like a load balancer for example Nginx for the API.


## Setup
Whether deployed as a container or not; setup will be the same.

Configuration is handled through YAML; depending on your setup either one or two files will need to be created, since a single file can contain the config for a producer and consumer. A example config can be found in the repo called: `example.yaml`

By default when run the app will look for a file called `config.yaml`. Although can be defined via two environment variables: `IO_CONSUMER_CONFIG` and `IO_PRODUCER_CONFIG`.


## With Docker
### Requirements
- RabbitMQ >= 3.11
- Docker


## Without Docker
### Requirements
- libvips installed
- go >= 1.20
- RabbitMQ >= 3.11

### Build
1. Run `go build`
2. Copy built binary `image-optimizer` to permanent location e.g. ~/.local/bin
3. Run
