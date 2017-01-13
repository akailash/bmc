#!/bin/sh
set -e # exit on an error

ERROR(){
    /bin/echo -e "\e[101m\e[97m[ERROR]\e[49m\e[39m $@"
}

WARNING(){
    /bin/echo -e "\e[101m\e[97m[WARNING]\e[49m\e[39m $@"
}

INFO(){
    /bin/echo -e "\e[104m\e[97m[INFO]\e[49m\e[39m $@"
}

exists() {
    type $1 > /dev/null 2>&1
}

exists docker || { ERROR "Please install docker (https://docs.docker.com/engine/installation/)"; exit 1; }
exists docker-compose || { ERROR "Please install docker-compose (https://docs.docker.com/compose/install/)"; exit 1; }
protoc --go_out=bmcaster bimodal.proto
protoc --go_out=bmcaster digest.proto
protoc --go_out=multicaster bimodal.proto
docker-compose stop #Remove if you need other docker containers running
docker-compose rm -f
docker-compose pull
INFO "Please run \`docker exec -it <node> bash\` in another terminal after the containers are up"
INFO "Please scale using command \`docker-compose scale multicaster=1 node=10\` to change the number of nodes"
docker-compose up --build 
