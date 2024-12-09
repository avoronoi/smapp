# Smapp
A microservice backend written in Go for a social media app.

The goal of this project is to better understand microservice architecture by building a scalable backend and deploying it to a cluster of virtual machines.

## Table of contents
- [Features](#features)
- [Technologies](#technologies)
- [Running the Application](#running-the-application)
  - [Running with Docker Compose](#running-with-docker-compose)
  - [Running with Docker Swarm](#running-with-docker-swarm)
  - [Applying Database Migrations](#applying-database-migrations)
- [Deploying on AWS](#deploying-on-aws)
  - [Deployment](#deployment)
  - [Application Load Balancer](#application-load-balancer)
  - [Security groups](#security-groups)
  - [IAM role](#iam-role)


## Features
- Signup and login
- Post creation, comment/like functionality and statistics
- Presigned links for the frontend to upload post images
- Following functionality and paginated feed

## Technologies
- Go, with the `net/http` package for request handling
- gRPC for inter-service communication
- MySQL for data storage
- Docker Swarm for container orchestration
- Amazon EC2 and Application Load Balancer for deployment
- Amazon S3 for image uploading and storage

## Running the Application

Whether you want to run the project locally or remotely, first clone the repository:
```bash
git clone https://github.com/avoronoi/smapp
```
Then, generate the Go code for the gRPC client and server from the `.proto` files. Ensure that you have Go and the Protocol Buffer compiler (`protoc`) installed. Additionally, install the following packages:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
Then execute:
```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative common/grpc/user/user.proto 
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative common/grpc/image/image.proto 
```
You will also need to install [gomplate](https://docs.gomplate.ca/installing/), a template renderer that will be used to generate the `docker-compose.yml` file.

There are two options to run the app: using Docker Compose and Docker Swarm.

### Running with Docker Compose
Docker Compose runs only on a single node, so it should **only be used for development**. It builds images from the services' `Dockerfile`s and supports [rebuilding services on code changes](https://docs.docker.com/compose/how-tos/file-watch/) using `docker compose watch`. The configurations for `docker compose watch` are specified in the `docker-compose.override.yml` file.

To run the application, first generate the `docker-compose.yml` file:
```bash
LOCAL=1 gomplate -f docker-compose.yml.tmpl -o docker-compose.yml
```
Next, create a MySQL root password and store it in the `secrets/mysql_password.txt` file. Then, create an RSA key pair and store it in the `secrets/jwt_private_key.pem` and `./jwt_public_key.pem` files.

You can customize these locations in the `docker-compose.override.yml` file.

### Running with Docker Swarm

#### Build and Push Images
Docker Swarm does not build images from `Dockerfile`s and requires prebuilt images to be stored in a registry.

You can use a container registry like Amazon ECR to store them. If you want to run Docker Swarm locally, you can create a container from Docker's `registry:2` image:
```bash
docker run -d -p 5000:5000 --name registry registry:2
```
And then use `localhost:5000` as registry.

Generate the `docker-compose.yml` file:
```bash
DEPLOY=1 gomplate -f docker-compose.yml.tmpl -o docker-compose.yml
```
You should also specify `LOCAL=1` if you intend to run Swarm locally.

Next, build and push images to the registry:
```bash
docker compose build --push
```

#### Run Images on a Swarm
First, initialize the Swarm on one of the nodes with `./deployment/init_swarm.sh`.
This command initializes the Swarm (making the current node its only member) and sets up Swarm secrets and configs.

Run `docker swarm join-token manager` or `docker swarm join-token worker` to get a command for joining the Swarm as a manager or worker, respectively. Execute the generated command on the nodes you want to add.

Some nodes act as gateways. When running on an EC2 cluster, these nodes should be the ones that the Application Load Balancer forwards requests to. To designate a node as a gateway, execute `./deployment/set_gateway.sh` on it.

Finally, deploy the application from one of the manager nodes:
```bash
./deployment/deploy.sh <REGISTRY_ENDPOINT>
```

### Applying Database Migrations

To apply database migrations, use the Flyway migration tool from within the appropriate container.

From the `user_db_migrations` container:
```bash
flyway -url=jdbc:mysql://user_db:3306/user_db?allowPublicKeyRetrieval=true -user=root -password=$(cat /run/secrets/mysql_password) migrate
```
From the `post_db_migrations` container:
```bash
flyway -url=jdbc:mysql://post_db:3306/post_db?allowPublicKeyRetrieval=true -user=root -password=$(cat /run/secrets/mysql_password) migrate
```

## Deploying on AWS

### Deployment

To install the necessary packages on EC2 instances before deploying the Swarm, run:
```bash
./deployment/install.sh <EC2_INSTANCE_IP> <SSH_PRIVATE_KEY_FILE>
```

Then follow [these instructions](#run-images-on-a-swarm) to deploy the Swarm.

### Application Load Balancer
Create an Application Load Balancer instance with a target group consisting of the gateway instances (the ones where `set_gateway.sh` was executed). Configure health checks for the target group to use the `/ping` path.

### Security Groups
The security group attached to the load balancer instance should allow incoming TCP traffic on port 80.

The security group attached to EC2 instances should allow following traffic:
- TCP on port 80 from the load balancer instance's security group
- TCP/UDP on port 7946, UDP on port 4789, TCP on port 2377 for Swarm inter-node communication
- TCP on port 3306 for MySQL

### IAM Role
The EC2 instances should have an IAM role attached with the `AmazonEC2ContainerRegistryFullAccess` and `AmazonS3FullAccess` policies.