#! /bin/bash

# Start the server, wait a bit, and provide a command line to run the client
./ruby-chain.rb & (sleep 5; /bin/bash)
