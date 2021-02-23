#!/bin/bash

docker run -d --name dev-php \
	-v "$(pwd)":/var \
	-p 80:80 \
	kjtully/otel-chain-dev-php

exit 0
