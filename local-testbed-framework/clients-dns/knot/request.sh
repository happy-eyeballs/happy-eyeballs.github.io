#!/bin/bash

dig +time=30 -p 15353 $1 @127.0.0.1 $2
