#!/bin/bash

docker-compose up -d --force-recreate
vault operator unseal
vault operator unseal
vault operator unseal

