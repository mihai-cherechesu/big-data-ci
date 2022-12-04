### Setup
Install `docker` and `docker-compose` on your local machine.

### Usage
Execute `./run.sh` and wait for the stack to be ready (it will take a few seconds as there are health checks for some services).

### Push changes
Doing some changes in either `parser` or `controller` and pushing them to the `main` branch will automatically build a multi-arch image, tag it with `:latest` and push it into a public registry.

This way, if you want to see how your new changes (or someone else's) behave when you have the full stack created, you just need to run the `run.sh` and everything that's on the `main` branch will be reflected in the containers.
