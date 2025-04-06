#!/bin/bash

# remove unused images to free space
docker image prune -a -f
